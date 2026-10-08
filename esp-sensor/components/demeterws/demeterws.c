#include "esp_websocket_client.h"
#include "esp_log.h"
#include <stdio.h>

static const char *TAG = "wsconn";

void send_sensor_data(esp_websocket_client_handle_t client, int target_type, int target_id, int measurement)
{
    char payload[50];
    sprintf(payload, "%d:%d:%d", target_type, target_id, measurement);
    ESP_LOGI(TAG, "Sending payload: %s", payload);

    int i = 0;
    while (i < 1) {
        if (esp_websocket_client_is_connected(client)) {
                esp_websocket_client_send_text(client, payload, strlen(payload), 0);
        }
        vTaskDelay(1000 / portTICK_PERIOD_MS);
    }

}

static void websocket_event_handler(void *handler_args, esp_event_base_t base, int32_t event_id, void *event_data)
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


esp_websocket_client_handle_t connect_websocket() {
  const esp_websocket_client_config_t ws_cfg = {
      .uri = CONFIG_WEBSOCKET_URL,
  };

  esp_websocket_client_handle_t client = esp_websocket_client_init(&ws_cfg);

  esp_websocket_register_events(client, WEBSOCKET_EVENT_ANY,
                                websocket_event_handler, (void *)client);
  esp_websocket_client_start(client);

  return client;
}
