#include "ds18b20-sensor.h"

#include "ds18b20.h"
#include "esp_log.h"
#include "onewire_bus.h"
#include "sdkconfig.h"

// Measurement type the server expects for water temperature.
#define DS18B20_TYPE_WATER_TEMPERATURE 7

static const char *TAG = "ds18b20-sensor";

static onewire_bus_handle_t s_bus;
static ds18b20_device_handle_t s_device;

static esp_err_t find_device(void)
{
    onewire_device_iter_handle_t iter = NULL;
    esp_err_t err = onewire_new_device_iter(s_bus, &iter);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to enumerate the 1-wire bus: %s", esp_err_to_name(err));
        return err;
    }

    onewire_device_t candidate;
    while ((err = onewire_device_iter_get_next(iter, &candidate)) == ESP_OK) {
        ds18b20_config_t config = {};
        ds18b20_device_handle_t device = NULL;

        if (ds18b20_new_device_from_enumeration(&candidate, &config, &device) != ESP_OK) {
            ESP_LOGI(TAG, "skipping unknown 1-wire device %016llX",
                     (unsigned long long)candidate.address);
            continue;
        }

        onewire_device_address_t address = 0;
        ds18b20_get_device_address(device, &address);
        ESP_LOGI(TAG, "found DS18B20 %016llX", (unsigned long long)address);
        s_device = device;
        break;
    }

    if (err != ESP_OK && err != ESP_ERR_NOT_FOUND) {
        ESP_LOGE(TAG, "1-wire enumeration failed: %s", esp_err_to_name(err));
    }
    onewire_del_device_iter(iter);

    if (s_device == NULL) {
        ESP_LOGE(TAG, "no DS18B20 found on the 1-wire bus");
        return ESP_ERR_NOT_FOUND;
    }
    return ESP_OK;
}

esp_err_t ds18b20_sensor_init(void)
{
    if (s_device != NULL) {
        return ESP_OK;
    }

    onewire_bus_config_t bus_config = {
        .bus_gpio_num = CONFIG_DS18B20_GPIO,
    };
    onewire_bus_rmt_config_t rmt_config = {
        .max_rx_bytes = 10, // 1 byte ROM command + 8 byte ROM + 1 byte command
    };

    esp_err_t err = onewire_new_bus_rmt(&bus_config, &rmt_config, &s_bus);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to create the 1-wire bus: %s", esp_err_to_name(err));
        return err;
    }

    return find_device();
}

size_t ds18b20_sensor_measure(void *ctx,
                              demeter_measurement_t *measurements,
                              size_t max_measurements)
{
    (void)ctx;

    if (measurements == NULL || max_measurements == 0) {
        return 0;
    }
    if (s_device == NULL) {
        ESP_LOGE(TAG, "measure called before init");
        return 0;
    }

    // Blocks for the conversion time, which is why this sensor is not on a
    // timer callback.
    esp_err_t err = ds18b20_trigger_temperature_conversion_for_all(s_bus);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to trigger a conversion: %s", esp_err_to_name(err));
        return 0;
    }

    float temperature = 0.0f;
    err = ds18b20_get_temperature(s_device, &temperature);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "failed to read the temperature: %s", esp_err_to_name(err));
        return 0;
    }

    ESP_LOGI(TAG, "temperature: %.2f C", temperature);

    measurements[0] = (demeter_measurement_t){
        .target_type = DS18B20_TYPE_WATER_TEMPERATURE,
        .target_id = CONFIG_DS18B20_SYSTEM_ID,
        .measurement = (int)(temperature * 1000.0f),
    };
    return 1;
}

demeter_sensor_t ds18b20_sensor(void)
{
    return (demeter_sensor_t){
        .name = TAG,
        .init = ds18b20_sensor_init,
        .measure = ds18b20_sensor_measure,
        .ctx = NULL,
        .interval_ms = CONFIG_SENSOR_INTERVAL_SECONDS * 1000U,
    };
}
