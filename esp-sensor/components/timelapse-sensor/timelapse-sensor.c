#include <stdio.h>
#include "timelapse-sensor.h"
#include "esp_camera.h"
#include "esp_camera_af.h"
#include "esp_wifi.h"
#include "esp_websocket_client.h"
#include "demeterws.h"

void func(void)
{

}

/*
https://github.com/espressif/esp32-camera/blob/master/examples/camera_example/main/take_picture.c
esp_camera_set_psram_mode()
CONFIG_CAMERA_PSRAM_DMA=true
*/

static const char *TIMELAPSE_TAG = "timelapse-sensor";

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
    .xclk_freq_hz = 20000000,
    .ledc_timer = LEDC_TIMER_0,
    .ledc_channel = LEDC_CHANNEL_0,

    .pixel_format = PIXFORMAT_JPEG, //YUV422,GRAYSCALE,RGB565,JPEG
    .frame_size = FRAMESIZE_QVGA,    //QQVGA-UXGA, For ESP32, do not use sizes above QVGA when not JPEG. The performance of the ESP32-S series has improved a lot, but JPEG mode always gives better frame rates.

    .jpeg_quality = 12, //0-63, for OV series camera sensors, lower number means higher quality
    .fb_count = 1,       //When jpeg mode is used, if fb_count more than one, the driver will work in continuous mode.
    .fb_location = CAMERA_FB_IN_PSRAM,
    .grab_mode = CAMERA_GRAB_WHEN_EMPTY,
    .jpeg_buffer_size = 0, //Set to a custom size in bytes for JPEG mode, or 0 to use the default size
};

static char *mac_address[12];

static esp_err_t init_camera(void)
{
    uint8_t baseMac[6];
    esp_err_t ret = esp_wifi_get_mac(WIFI_IF_STA, baseMac);
    if (ret != ESP_OK) {
        ESP_LOGE(TIMELAPSE_TAG, "failed to get esp MAC");
    }

    sprintf(mac_address, "%02x%02x%02x%02x%02x%02x", baseMac[0], baseMac[1], baseMac[2], baseMac[3], baseMac[4], baseMac[5]);

    //initialize the camera
    esp_err_t err = esp_camera_init(&camera_config);
    if (err != ESP_OK)
    {
        ESP_LOGE(TAG, "Camera Init Failed");
        return err;
    }

    return ESP_OK;
}

static void maybe_init_autofocus(void)
{
    sensor_t *s = esp_camera_sensor_get();
    if (!s) {
        ESP_LOGW(TAG, "AF: no sensor handle");
        return;
    }

    if (!esp_camera_af_is_supported(s)) {
        ESP_LOGI(TAG, "AF: not supported by this sensor");
        return;
    }

    esp_camera_af_config_t af_cfg = {
        .mode = ESP_CAMERA_AF_MODE_AUTO,
        .timeout_ms = CONFIG_CAMERA_AF_DEFAULT_TIMEOUT_MS,
    };

    esp_err_t ret = esp_camera_af_init(s, &af_cfg);
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "AF init failed: %s", esp_err_to_name(ret));
        return;
    }

    ESP_LOGI(TAG, "AF initialized (AUTO mode)");
}


void send_sensor_data(esp_websocket_client_handle_t client, int target_type, int target_id, int measurement)
{
    char payload[50];
    sprintf(payload, "%d:%d:%d", target_type, target_id, measurement);
    ESP_LOGI(TAG, "Sending payload: %s", payload);

    int i = 0;
    while (i < 1) {
        if (esp_websocket_client_is_connected(client)) {
                esp_websocket_client_send_text(client, payload, strlen(payload), 0);
                i += 1;
        }
        vTaskDelay(1000 / portTICK_PERIOD_MS);
    }

}


void send_photo(esp_webocket_client_handle_t client) {
    if (ESP_OK != init_camera()) {
        ESP_LOGE(TIMELAPSE_TAG, "failed to  init camera");
        return;
    }

    camera_fb_t *pic = esp_camera_fb_get();

    ESP_LOGI(TIMELAPSE_TAG, "took picture of size %zu bytes", pic->len);


    int i = 0;
    while (i < 1) {
        if (esp_websocket_client_is_connected(client)) {
                esp_websocket_client_send_bin_partial(client, mac_address, 12, 0);
                esp_websocket_client_send_cont_msg(client, pic->buf, pic->len, 0);
                esp_websocket_client_send_fin(client, 0);
                i += 1;
        }
        vTaskDelay(1000 / portTICK_PERIOD_MS);
    }
    // pic->buf = image
}
