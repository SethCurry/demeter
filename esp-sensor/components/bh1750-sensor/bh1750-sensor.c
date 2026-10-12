#include "bh1750-sensor.h"
#include "bh1750.h"
#include "esp_websocket_client.h"
#include "demeterws.h"
#include <esp_err.h>
#include "esp_log.h"

static const char *BH_TAG = "bh1750-sensor";

static i2c_dev_t dev;


void send_bh1750_data(esp_websocket_client_handle_t client, int target_id) {
    uint16_t lux;

    if (bh1750_read(&dev, &lux) != ESP_OK) {
        ESP_LOGE(BH_TAG, "bh1750 failed\n");
    } else {
      ESP_LOGI(BH_TAG, "bh1750 measurement: lux: %d\n", lux);
      send_sensor_data(client, 10, target_id, lux);
    }
}

void get_bh1750_dev() {
    memset(&dev, 0, sizeof(i2c_dev_t)); // Zero descriptor

  ESP_ERROR_CHECK(bh1750_init_desc(&dev, BH1750_ADDR_LO, 0, CONFIG_I2C_DATA_GPIO, CONFIG_I2C_CLOCK_GPIO));
                ESP_ERROR_CHECK(bh1750_setup(&dev, BH1750_MODE_CONTINUOUS,
                                             BH1750_RES_HIGH));
}
