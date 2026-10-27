// Periodic sensor publisher.
//
// Every configured sensor exposes a demeter_sensor_t -- an optional init hook,
// a measure hook and the period it wants to run at. app_main registers those
// descriptors with a single orchestrator_t and then just runs it; the
// orchestrator decides which sensors are due, measures them and sleeps until
// the next one. app_main is already a FreeRTOS task, so this behaves like a set
// of timers without the cost of the alternatives:
//   - no per-sensor task (saves a ~3.5 KB stack + TCB per sensor),
//   - no esp_timer/Tmr Svc callback (those must be quick and non-blocking;
//     ds18b20 conversions, camera captures and the websocket send path block
//     for seconds).

#include "esp_err.h"
#include "esp_log.h"
#include "esp_websocket_client.h"
#include "nvs_flash.h"
#include "sdkconfig.h"

#include "demeter.h"
#include "wifi.h"

#if CONFIG_I2C_DATA_GPIO != 0
#include "i2cdev.h"
#endif

#if CONFIG_DS18B20_GPIO != 0
#include "ds18b20-sensor.h"
#endif

#if CONFIG_SHT_ENCLOSURE_ID != 0
#include "sht4x-sensor.h"
#endif

#if CONFIG_BH_PLANT_SITE_ID != 0
#include "bh1750-sensor.h"
#endif

#include "timelapse-sensor.h"

static const char *TAG = "demeter";

static esp_websocket_client_handle_t s_ws;
static orchestrator_t s_orchestrator;

// A sensor that fails to start (probe unplugged, camera missing, ...) is
// skipped so the remaining sensors keep publishing.
static void add_sensor(demeter_sensor_t sensor)
{
    esp_err_t err = orchestrator_add_sensor(&s_orchestrator, sensor);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "skipping sensor %s: %s",
                 sensor.name != NULL ? sensor.name : "unnamed",
                 esp_err_to_name(err));
    }
}

// register_sensors() is the only place that decides which sensors exist, so
// app_main stays free of #ifdefs.
static void register_sensors(void)
{
    ESP_ERROR_CHECK(orchestrator_init(&s_orchestrator));

#if CONFIG_I2C_DATA_GPIO != 0
    ESP_ERROR_CHECK(i2cdev_init());
#endif

#if CONFIG_DS18B20_GPIO != 0
    add_sensor(ds18b20_sensor());
#endif
#if CONFIG_SHT_ENCLOSURE_ID != 0
    add_sensor(sht4x_sensor());
#endif
#if CONFIG_BH_PLANT_SITE_ID != 0
    add_sensor(bh1750_sensor());
#endif
    // The camera streams pictures over the websocket, so it needs the client.
    add_sensor(timelapse_sensor(s_ws));
}

void app_main(void)
{
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    start_wifi();
    s_ws = connect_websocket();
    register_sensors();

    for (;;) {
        orchestrator_run(&s_orchestrator, s_ws);
    }
}
