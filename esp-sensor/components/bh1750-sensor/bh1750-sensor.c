#include "bh1750-sensor.h"

#include <stdbool.h>
#include <string.h>

#include "bh1750.h"
#include "esp_log.h"
#include "sdkconfig.h"

// Measurement type the server expects for illuminance.
#define BH1750_TYPE_LUX 10

static const char *TAG = "bh1750-sensor";

static i2c_dev_t s_dev;
static bool s_ready;

esp_err_t bh1750_sensor_init(void)
{
    if (s_ready) {
        return ESP_OK;
    }

    memset(&s_dev, 0, sizeof(s_dev));

    esp_err_t err = bh1750_init_desc(&s_dev, BH1750_ADDR_LO, 0,
                                     CONFIG_I2C_DATA_GPIO,
                                     CONFIG_I2C_CLOCK_GPIO);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to describe the device: %s", esp_err_to_name(err));
        return err;
    }

    err = bh1750_setup(&s_dev, BH1750_MODE_CONTINUOUS, BH1750_RES_HIGH);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to set the device up: %s", esp_err_to_name(err));
        return err;
    }

    s_ready = true;
    return ESP_OK;
}

size_t bh1750_sensor_measure(void *ctx,
                             demeter_measurement_t *measurements,
                             size_t max_measurements)
{
    (void)ctx;

    if (measurements == NULL || max_measurements == 0) {
        return 0;
    }
    if (!s_ready) {
        ESP_LOGE(TAG, "measure called before init");
        return 0;
    }

    uint16_t lux = 0;
    esp_err_t err = bh1750_read(&s_dev, &lux);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to read the illuminance: %s", esp_err_to_name(err));
        return 0;
    }

    ESP_LOGI(TAG, "illuminance: %u lux", (unsigned)lux);

    measurements[0] = (demeter_measurement_t){
        .target_type = BH1750_TYPE_LUX,
        .target_id = CONFIG_BH_PLANT_SITE_ID,
        .measurement = (int)lux,
    };
    return 1;
}

demeter_sensor_t bh1750_sensor(void)
{
    return (demeter_sensor_t){
        .name = TAG,
        .init = bh1750_sensor_init,
        .measure = bh1750_sensor_measure,
        .ctx = NULL,
        .interval_ms = CONFIG_SENSOR_INTERVAL_SECONDS * 1000U,
    };
}
