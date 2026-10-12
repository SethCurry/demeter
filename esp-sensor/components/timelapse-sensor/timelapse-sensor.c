#include <stdint.h>
#include <stdio.h>
#include "timelapse-sensor.h"
#include "esp_camera.h"
#include "esp_camera_af.h"
#include "esp_wifi.h"
#include "esp_websocket_client.h"
#include "demeterws.h"
#include "esp_log.h"

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

static char mac_address[13];


esp_err_t init_camera(void)
{
    uint8_t baseMac[6];
    esp_err_t ret = esp_wifi_get_mac(WIFI_IF_STA, baseMac);
    if (ret != ESP_OK) {
        ESP_LOGE(TIMELAPSE_TAG, "failed to get esp MAC");
    }

    sprintf(&mac_address[0], "%02x%02x%02x%02x%02x%02x", baseMac[0], baseMac[1], baseMac[2], baseMac[3], baseMac[4], baseMac[5]);



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
    camera_fb_t *pic = esp_camera_fb_get();
    int i = 0;

    while (i < 10) {
        if (pic == NULL) {
            vTaskDelay(1000 / portTICK_PERIOD_MS);
            pic = esp_camera_fb_get();
            i++;
        } else {
            i = 11;
        }
    }

    if (pic == NULL) {
        ESP_LOGI(TIMELAPSE_TAG, "failed to take picture");
        return;
    }

    ESP_LOGI(TIMELAPSE_TAG, "took picture of size %zu bytes", pic->len);


    i = 0;

    uint8_t *starting_address = pic->buf;

    esp_websocket_client_send_bin_partial(client, mac_address, 12, 0);

    while (i < pic->len) {
        if (esp_websocket_client_is_connected(client)) {
                int remaining = (pic->len*sizeof (uint8_t)) - i;
                int to_write = 512;
                if (remaining < 512) {
                    to_write = remaining;
                }
                esp_websocket_client_send_cont_msg(client, (char *)starting_address, to_write, 0);
                starting_address = starting_address + 512;//*sizeof (uint8_t);
                i += to_write;
                vTaskDelay(50 / portTICK_PERIOD_MS);
        } else {
        vTaskDelay(100 / portTICK_PERIOD_MS);
        }
    }
    esp_websocket_client_send_fin(client, 0);
    esp_camera_fb_return(pic);
}
