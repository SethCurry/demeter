#include "sht4x-sensor.h"

#include <stdbool.h>
#include <string.h>

#include "esp_log.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "sdkconfig.h"
#include "sht4x.h"

// Measurement types the server expects.
#define SHT4X_TYPE_TEMPERATURE 2
#define SHT4X_TYPE_HUMIDITY 3

static const char *TAG = "sht4x-sensor";

static sht4x_t s_dev;
static bool s_ready;

esp_err_t sht4x_sensor_init(void)
{
    if (s_ready) {
        return ESP_OK;
    }

    memset(&s_dev, 0, sizeof(s_dev));

    esp_err_t err = sht4x_init_desc(&s_dev, 0, CONFIG_I2C_DATA_GPIO,
                                    CONFIG_I2C_CLOCK_GPIO);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to describe the device: %s", esp_err_to_name(err));
        return err;
    }

    err = sht4x_init(&s_dev);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to initialise the device: %s", esp_err_to_name(err));
        return err;
    }

    s_ready = true;
    return ESP_OK;
}

size_t sht4x_sensor_measure(void *ctx,
                            demeter_measurement_t *measurements,
                            size_t max_measurements)
{
    (void)ctx;

    if (measurements == NULL || max_measurements < 2) {
        return 0;
    }
    if (!s_ready) {
        ESP_LOGE(TAG, "measure called before init");
        return 0;
    }

    esp_err_t err = sht4x_start_measurement(&s_dev);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to start a measurement: %s", esp_err_to_name(err));
        return 0;
    }
    // The conversion time is reported in ticks.
    vTaskDelay(sht4x_get_measurement_duration(&s_dev));

    float temperature = 0.0f;
    float humidity = 0.0f;
    err = sht4x_get_results(&s_dev, &temperature, &humidity);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to read the results: %s", esp_err_to_name(err));
        return 0;
    }

    ESP_LOGI(TAG, "temperature: %.2f C, humidity: %.2f %%", temperature, humidity);

    measurements[0] = (demeter_measurement_t){
        .target_type = SHT4X_TYPE_TEMPERATURE,
        .target_id = CONFIG_SHT_ENCLOSURE_ID,
        .measurement = (int)(temperature * 1000.0f),
    };
    measurements[1] = (demeter_measurement_t){
        .target_type = SHT4X_TYPE_HUMIDITY,
        .target_id = CONFIG_SHT_ENCLOSURE_ID,
        .measurement = (int)(humidity * 1000.0f),
    };
    return 2;
}

demeter_sensor_t sht4x_sensor(void)
{
    return (demeter_sensor_t){
        .name = TAG,
        .init = sht4x_sensor_init,
        .measure = sht4x_sensor_measure,
        .ctx = NULL,
        .interval_ms = CONFIG_SENSOR_INTERVAL_SECONDS * 1000U,
    };
}
