#include "timelapse-sensor.h"

#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "esp_camera.h"
#include "esp_camera_af.h"
#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_websocket_client.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

/*
https://github.com/espressif/esp32-camera/blob/master/examples/camera_example/main/take_picture.c
esp_camera_set_psram_mode()
CONFIG_CAMERA_PSRAM_DMA=true
*/

static const char *TIMELAPSE_TAG = "timelapse-sensor";

// Largest frame handed to a single continuation message.
#define TIMELAPSE_CHUNK_SIZE 512
// How long to wait for a frame from the camera before giving up.
#define TIMELAPSE_CAPTURE_RETRIES 10
// How long the upload may stall on a disconnected websocket before the picture
// is abandoned, so the orchestrator is not blocked forever.
#define TIMELAPSE_UPLOAD_TIMEOUT_MS 30000

static camera_config_t camera_config = {
    .pin_pwdn = CAM_PIN_PWDN,
    .pin_reset = CAM_PIN_RESET,
    .pin_xclk = CAM_PIN_XCLK,
    .pin_sccb_sda = CAM_PIN_SIOD,
    .pin_sccb_scl = CAM_PIN_SIOC,

    .pin_d7 = CAM_PIN_D7,
    .pin_d6 = CAM_PIN_D6,
    .pin_d5 = CAM_PIN_D5,
    .pin_d4 = CAM_PIN_D4,
    .pin_d3 = CAM_PIN_D3,
    .pin_d2 = CAM_PIN_D2,
    .pin_d1 = CAM_PIN_D1,
    .pin_d0 = CAM_PIN_D0,
    .pin_vsync = CAM_PIN_VSYNC,
    .pin_href = CAM_PIN_HREF,
    .pin_pclk = CAM_PIN_PCLK,

    //XCLK 20MHz or 10MHz for OV2640 double FPS (Experimental)
    .xclk_freq_hz = 5000000,
    .ledc_timer = LEDC_TIMER_0,
    .ledc_channel = LEDC_CHANNEL_0,

    .pixel_format = PIXFORMAT_JPEG, //YUV422,GRAYSCALE,RGB565,JPEG
    .frame_size = FRAMESIZE_5MP,    //QQVGA-UXGA, For ESP32, do not use sizes above QVGA when not JPEG. The performance of the ESP32-S series has improved a lot, but JPEG mode always gives better frame rates.

    .jpeg_quality = 12, //0-63, for OV series camera sensors, lower number means higher quality
    .fb_count = 1,       //When jpeg mode is used, if fb_count more than one, the driver will work in continuous mode.
    .fb_location = CAMERA_FB_IN_PSRAM,
    .grab_mode = CAMERA_GRAB_LATEST,
    .jpeg_buffer_size = 0, //Set to a custom size in bytes for JPEG mode, or 0 to use the default size
};

// The server identifies the camera by the station MAC address.
static char mac_address[13];

esp_err_t init_camera(void)
{
    uint8_t baseMac[6] = { 0 };
    esp_err_t ret = esp_wifi_get_mac(WIFI_IF_STA, baseMac);
    if (ret != ESP_OK) {
        ESP_LOGE(TIMELAPSE_TAG, "failed to get esp MAC");
        return ret;
    }

    snprintf(mac_address, sizeof(mac_address), "%02x%02x%02x%02x%02x%02x",
             baseMac[0], baseMac[1], baseMac[2], baseMac[3], baseMac[4], baseMac[5]);

    //initialize the camera
    esp_err_t err = esp_camera_init(&camera_config);
    if (err != ESP_OK)
    {
        ESP_LOGE(TIMELAPSE_TAG, "Camera Init Failed");
        return err;
    }

    return ESP_OK;
}

void maybe_init_autofocus(void)
{
    sensor_t *s = esp_camera_sensor_get();
    if (!s) {
        ESP_LOGW(TIMELAPSE_TAG, "AF: no sensor handle");
        return;
    }

    if (!esp_camera_af_is_supported(s)) {
        ESP_LOGI(TIMELAPSE_TAG, "AF: not supported by this sensor");
        return;
    }

    esp_camera_af_config_t af_cfg = {
        .mode = ESP_CAMERA_AF_MODE_AUTO,
        .timeout_ms = 30000,
    };

    esp_err_t ret = esp_camera_af_init(s, &af_cfg);
    if (ret != ESP_OK) {
        ESP_LOGW(TIMELAPSE_TAG, "AF init failed: %s", esp_err_to_name(ret));
        return;
    }

    ESP_LOGI(TIMELAPSE_TAG, "AF initialized (AUTO mode)");
}


void send_demeter_timelapse(esp_websocket_client_handle_t client) {
    if (client == NULL) {
        ESP_LOGW(TIMELAPSE_TAG, "no websocket client, skipping the timelapse");
        return;
    }

    camera_fb_t *pic = NULL;
    for (int attempt = 0; attempt < TIMELAPSE_CAPTURE_RETRIES; attempt++) {
        pic = esp_camera_fb_get();
        if (pic != NULL) {
            break;
        }
        vTaskDelay(pdMS_TO_TICKS(1000));
    }

    if (pic == NULL) {
        ESP_LOGI(TIMELAPSE_TAG, "failed to take picture");
        return;
    }

    ESP_LOGI(TIMELAPSE_TAG, "took picture of size %zu bytes", pic->len);

    size_t sent = 0;
    uint8_t *chunk = pic->buf;

    esp_websocket_client_send_bin_partial(client, mac_address,
                                          strlen(mac_address), 0);

    TickType_t deadline = xTaskGetTickCount()
                          + pdMS_TO_TICKS(TIMELAPSE_UPLOAD_TIMEOUT_MS);

    while (sent < pic->len) {
        if (!esp_websocket_client_is_connected(client)) {
            if ((int32_t)(xTaskGetTickCount() - deadline) >= 0) {
                ESP_LOGW(TIMELAPSE_TAG, "upload timed out after %zu of %zu bytes",
                         sent, pic->len);
                break;
            }
            vTaskDelay(pdMS_TO_TICKS(100));
            continue;
        }

        size_t remaining = pic->len - sent;
        size_t to_write = remaining < TIMELAPSE_CHUNK_SIZE ? remaining
                                                           : TIMELAPSE_CHUNK_SIZE;
        esp_websocket_client_send_cont_msg(client, (char *)chunk, to_write, 0);
        chunk += to_write;
        sent += to_write;
        vTaskDelay(pdMS_TO_TICKS(50));
    }

    if (sent >= pic->len) {
        esp_websocket_client_send_fin(client, 0);
    }

    esp_camera_fb_return(pic);
}

esp_err_t timelapse_sensor_init(void)
{
    esp_err_t err = init_camera();
    if (err != ESP_OK) {
        return err;
    }

    maybe_init_autofocus();
    return ESP_OK;
}

size_t timelapse_sensor_measure(void *ctx,
                                demeter_measurement_t *measurements,
                                size_t max_measurements)
{
    (void)measurements;
    (void)max_measurements;

    // The picture is streamed over the websocket instead of being reported as
    // a scalar measurement, so nothing is returned to the orchestrator.
    send_demeter_timelapse((esp_websocket_client_handle_t)ctx);
    return 0;
}

demeter_sensor_t timelapse_sensor(esp_websocket_client_handle_t client)
{
    return (demeter_sensor_t){
        .name = TIMELAPSE_TAG,
        .init = timelapse_sensor_init,
        .measure = timelapse_sensor_measure,
        .ctx = client,
        .interval_ms = TIMELAPSE_SENSOR_INTERVAL_MS,
    };
}
