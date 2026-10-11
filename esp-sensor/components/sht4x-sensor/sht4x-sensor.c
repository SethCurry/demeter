#include <stdio.h>
#include "sht4x-sensor.h"
#include "sht4x.h"
#include "esp_websocket_client.h"
#include "demeterws.h"
#include <esp_err.h>
#include "esp_log.h"

const char *TAG = "sht4x-sensor";

static sht4x_t dev;


void send_sht_data(esp_websocket_client_handle_t client, int target_id,
                   sht4x_t *some) {

  uint8_t duration = sht4x_get_measurement_duration(&dev);
  ESP_ERROR_CHECK(sht4x_start_measurement(&dev));
  vTaskDelay(duration);
  float temperature;
  float humidity;
  ESP_ERROR_CHECK(sht4x_get_results(&dev, &temperature, &humidity));
  ESP_LOGI(TAG, "sht4x measurement: temperature: %.2f, humidity: %.2f", temperature, humidity);
  send_sensor_data(client, 2, target_id, temperature*1000.0);
  send_sensor_data(client, 3, target_id, humidity*1000.0);
}

sht4x_t get_shtx_dev() {
    ESP_ERROR_CHECK(i2cdev_init());
    memset(&dev, 0, sizeof(sht4x_t));

    ESP_ERROR_CHECK(sht4x_init_desc(&dev, 0, CONFIG_SHT_DATA_GPIO, CONFIG_SHT_CLOCK_GPIO));
    ESP_ERROR_CHECK(sht4x_init(&dev));

    return dev;
}
