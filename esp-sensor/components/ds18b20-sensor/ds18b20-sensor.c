#include "onewire_bus.h"
#include "sdkconfig.h"
#include "esp_log.h"
#include "ds18b20.h"
#include "esp_websocket_client.h"
#include "demeterws.h"
#include "ds18b20-sensor.h"

const char *TAG = "ds18b20-sensor";

onewire_bus_handle_t get_bus_handle() {
    onewire_bus_handle_t bus = NULL;
    onewire_bus_config_t bus_config = {
        .bus_gpio_num = CONFIG_DS18B20_GPIO,
    };

    onewire_bus_rmt_config_t rmt_config = {
            .max_rx_bytes = 10, // 1byte ROM command + 8byte ROM number + 1byte device command
    };
    ESP_ERROR_CHECK(onewire_new_bus_rmt(&bus_config, &rmt_config, &bus));

    return bus;
}

ds18b20_device_handle_t get_ds18b20_handle(onewire_bus_handle_t bus) {
    int ds18b20_device_num = 0;
    ds18b20_device_handle_t ds18b20s[1];
    onewire_device_iter_handle_t iter = NULL;
    onewire_device_t next_onewire_device;
    esp_err_t search_result = ESP_OK;

    // create 1-wire device iterator, which is used for device search
    ESP_ERROR_CHECK(onewire_new_device_iter(bus, &iter));
    ESP_LOGI(TAG, "Device iterator created, start searching...");
    do {
        search_result = onewire_device_iter_get_next(iter, &next_onewire_device);
        if (search_result == ESP_OK) { // found a new device, let's check if we can upgrade it to a DS18B20
            ds18b20_config_t ds_cfg = {};
            onewire_device_address_t address;
            // check if the device is a DS18B20, if so, return the ds18b20 handle
            if (ds18b20_new_device_from_enumeration(&next_onewire_device, &ds_cfg, &ds18b20s[ds18b20_device_num]) == ESP_OK) {
                ds18b20_get_device_address(ds18b20s[ds18b20_device_num], &address);
                ESP_LOGI(TAG, "Found a DS18B20[%d], address: %016llX", ds18b20_device_num, address);
                ds18b20_device_num++;
            } else {
                ESP_LOGI(TAG, "Found an unknown device, address: %016llX", next_onewire_device.address);
            }
                }
    } while (search_result != ESP_ERR_NOT_FOUND);
    ESP_ERROR_CHECK(onewire_del_device_iter(iter));
    ESP_LOGI(TAG, "Searching done, %d DS18B20 device(s) found", ds18b20_device_num);

    ds18b20_trigger_temperature_conversion_for_all(bus);


    ds18b20_device_handle_t ds18b20_handle = ds18b20s[0];

    return ds18b20_handle;
}

void send_dsb_data(esp_websocket_client_handle_t client, int target_id, dsb_handle_t *dsb) {
    float temperature;
    int get_result = ds18b20_get_temperature(dsb->ds18b20, &temperature);

    if (get_result != ESP_OK) {
        ESP_LOGE(TAG, "Failed to get temperature: %d", get_result);
        return;
    } else {
        ESP_LOGI(TAG, "Temperature: %.2f", temperature);
    }

    send_sensor_data(client, 7, target_id, temperature*1000.0);
}
