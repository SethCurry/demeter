#include "demeter.h"

#include <inttypes.h>
#include <stdio.h>
#include <string.h>

#include "esp_event.h"
#include "esp_log.h"
#include "freertos/task.h"
#include "sdkconfig.h"

// How long a measurement waits for the websocket to (re)connect before it is
// dropped, and how long a single text frame may take to go out.
#define DEMETER_CONNECT_TIMEOUT_MS 10000
#define DEMETER_SEND_TIMEOUT_MS 1000
// Fallback period when descriptor leaves it at 0.
#define DEMETER_DEFAULT_INTERVAL_MS 1000

static const char *TAG = "demeter";

static const char *sensor_name(const demeter_sensor_t *sensor)
{
    return sensor->name != NULL ? sensor->name : "unnamed";
}

void send_sensor_data(esp_websocket_client_handle_t client, int target_type,
                      int target_id, int measurement)
{
    if (client == NULL) {
        ESP_LOGW(TAG, "no websocket client, dropping %d:%d:%d", target_type,
                 target_id, measurement);
        return;
    }

    char payload[48];
    int len = snprintf(payload, sizeof(payload), "%d:%d:%d", target_type,
                       target_id, measurement);
    if (len <= 0 || (size_t)len >= sizeof(payload)) {
        ESP_LOGE(TAG, "measurement does not fit the wire format");
        return;
    }

    // The client reconnects on its own, so wait for it -- but bounded, because
    // this runs on the orchestrator task that also feeds every other sensor.
    TickType_t deadline = xTaskGetTickCount()
                          + pdMS_TO_TICKS(DEMETER_CONNECT_TIMEOUT_MS);
    while (!esp_websocket_client_is_connected(client)) {
        if ((int32_t)(xTaskGetTickCount() - deadline) >= 0) {
            ESP_LOGW(TAG, "websocket not connected, dropping payload %s", payload);
            return;
        }
        vTaskDelay(pdMS_TO_TICKS(100));
    }

    ESP_LOGI(TAG, "Sending payload: %s", payload);
    if (esp_websocket_client_send_text(client, payload, len,
                                      pdMS_TO_TICKS(DEMETER_SEND_TIMEOUT_MS)) < 0) {
        ESP_LOGE(TAG, "failed to send payload %s", payload);
    }
}

size_t measure_sensor(esp_websocket_client_handle_t client,
                      const demeter_sensor_t *sensor)
{
    if (sensor == NULL || sensor->measure == NULL) {
        return 0;
    }

    demeter_measurement_t measurements[DEMETER_MAX_MEASUREMENTS];
    size_t count = sensor->measure(sensor->ctx, measurements,
                                   DEMETER_MAX_MEASUREMENTS);
    if (count > DEMETER_MAX_MEASUREMENTS) {
        count = DEMETER_MAX_MEASUREMENTS;
    }

    for (size_t i = 0; i < count; i++) {
        send_sensor_data(client, measurements[i].target_type,
                         measurements[i].target_id, measurements[i].measurement);
    }

    return count;
}

esp_err_t orchestrator_init(orchestrator_t *orchestrator)
{
    if (orchestrator == NULL) {
        return ESP_ERR_INVALID_ARG;
    }

    memset(orchestrator, 0, sizeof(*orchestrator));
    return ESP_OK;
}

esp_err_t orchestrator_add_sensor(orchestrator_t *orchestrator,
                                  demeter_sensor_t sensor)
{
    if (orchestrator == NULL || sensor.measure == NULL) {
        return ESP_ERR_INVALID_ARG;
    }
    if (orchestrator->num_sensors >= DEMETER_MAX_SENSORS) {
        ESP_LOGE(TAG, "orchestrator already holds %d sensors, dropping %s",
                 DEMETER_MAX_SENSORS, sensor_name(&sensor));
        return ESP_ERR_NO_MEM;
    }

    if (sensor.interval_ms == 0) {
        ESP_LOGW(TAG, "sensor %s has no interval, defaulting to %d ms",
                 sensor_name(&sensor), DEMETER_DEFAULT_INTERVAL_MS);
        sensor.interval_ms = DEMETER_DEFAULT_INTERVAL_MS;
    }

    if (sensor.init != NULL) {
        esp_err_t err = sensor.init();
        if (err != ESP_OK) {
            ESP_LOGE(TAG, "starting sensor %s failed: %s", sensor_name(&sensor),
                     esp_err_to_name(err));
            return err;
        }
    }

    orchestrator->slots[orchestrator->num_sensors++] = (demeter_sensor_slot_t){
        .sensor = sensor,
        .next_run = xTaskGetTickCount(),  // due on the first pass
    };

    ESP_LOGI(TAG, "registered sensor %s, every %" PRIu32 " ms",
             sensor_name(&sensor), sensor.interval_ms);
    return ESP_OK;
}

void orchestrator_measure(orchestrator_t *orchestrator,
                          esp_websocket_client_handle_t client)
{
    if (orchestrator == NULL) {
        return;
    }

    TickType_t now = xTaskGetTickCount();
    for (size_t i = 0; i < orchestrator->num_sensors; i++) {
        demeter_sensor_slot_t *slot = &orchestrator->slots[i];

        // Signed difference so the comparison still works when the tick
        // counter wraps around.
        if ((int32_t)(now - slot->next_run) < 0) {
            continue;
        }

        measure_sensor(client, &slot->sensor);

        TickType_t period = pdMS_TO_TICKS(slot->sensor.interval_ms);
        if (period == 0) {
            period = 1;
        }
        // Schedule from completion time, so a slow sensor (camera, 1-wire
        // conversion) does not immediately become due again.
        slot->next_run = xTaskGetTickCount() + period;
    }
}

static TickType_t orchestrator_ticks_until_next(const orchestrator_t *orchestrator)
{
    TickType_t now = xTaskGetTickCount();
    TickType_t best = 0;
    size_t found = 0;

    for (size_t i = 0; i < orchestrator->num_sensors; i++) {
        int32_t delta = (int32_t)(orchestrator->slots[i].next_run - now);
        if (delta <= 0) {
            return 0;
        }
        if (!found || (TickType_t)delta < best) {
            best = (TickType_t)delta;
            found = 1;
        }
    }

    return found ? best : pdMS_TO_TICKS(DEMETER_DEFAULT_INTERVAL_MS);
}

void orchestrator_run(orchestrator_t *orchestrator,
                      esp_websocket_client_handle_t client)
{
    if (orchestrator == NULL) {
        vTaskDelay(pdMS_TO_TICKS(DEMETER_DEFAULT_INTERVAL_MS));
        return;
    }

    orchestrator_measure(orchestrator, client);

    TickType_t delay = orchestrator_ticks_until_next(orchestrator);
    vTaskDelay(delay > 0 ? delay : 1);
}

static void websocket_event_handler(void *handler_args, esp_event_base_t base,
                                    int32_t event_id, void *event_data)
{
    esp_websocket_event_data_t *data = (esp_websocket_event_data_t *)event_data;
    switch (event_id) {
    case WEBSOCKET_EVENT_BEGIN:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_BEGIN");
        break;
#if WS_TRANSPORT_HEADER_CALLBACK_SUPPORT
    case WEBSOCKET_EVENT_HEADER_RECEIVED:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_HEADER_RECEIVED: %.*s", data->data_len, data->data_ptr);
        break;
#endif
    case WEBSOCKET_EVENT_CONNECTED:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_CONNECTED");
        break;
    case WEBSOCKET_EVENT_DISCONNECTED:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_DISCONNECTED");
        break;
    case WEBSOCKET_EVENT_DATA:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_DATA");
        ESP_LOGI(TAG, "Received opcode=%d", data->op_code);
        if (data->op_code == 0x08 && data->data_len == 2) {
            ESP_LOGW(TAG, "Received closed message with code=%d", 256 * data->data_ptr[0] + data->data_ptr[1]);
        } else {
            ESP_LOGW(TAG, "Received=%.*s", data->data_len, (char *)data->data_ptr);
        }

        // If received data contains json structure it succeed to parse
        ESP_LOGW(TAG, "Total payload length=%d, data_len=%d, current payload offset=%d\r\n", data->payload_len, data->data_len, data->payload_offset);

        break;
    case WEBSOCKET_EVENT_ERROR:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_ERROR");
        break;
    case WEBSOCKET_EVENT_FINISH:
        ESP_LOGI(TAG, "WEBSOCKET_EVENT_FINISH");
        break;
    }
}


esp_websocket_client_handle_t connect_websocket(void)
{
    const esp_websocket_client_config_t ws_cfg = {
        .uri = CONFIG_WEBSOCKET_URL,
        .task_core_id = tskNO_AFFINITY,
    };

    esp_websocket_client_handle_t client = esp_websocket_client_init(&ws_cfg);
    if (client == NULL) {
        ESP_LOGE(TAG, "failed to create the websocket client");
        return NULL;
    }

    esp_websocket_register_events(client, WEBSOCKET_EVENT_ANY,
                                  websocket_event_handler, (void *)client);
    esp_websocket_client_start(client);

    return client;
}
