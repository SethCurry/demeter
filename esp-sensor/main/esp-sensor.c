// Periodic sensor publisher.
//
// Each configured sensor registers a "job" -- a function plus a period --
// in one list. app_main (already a FreeRTOS task) runs the list, so every
// job acts like a timer without the costs of the alternatives:
//   - no per-sensor task (saves a ~3.5 KB stack + TCB per job),
//   - no esp_timer/Tmr Svc callback (those must be quick and non-blocking;
//     ds18b20 conversions and the websocket send path block for seconds).

#include <stdint.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "sdkconfig.h"
#include "wifi.h"
#include "demeterws.h"
#include "nvs_flash.h"

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

#define SENSOR_PERIOD_MS CONFIG_SENSOR_INTERVAL_SECONDS*1000
#define MAX_JOBS 10

typedef void (*job_fn_t)(void *arg);

typedef struct {
    job_fn_t fn;          // runs to completion when due
    void *arg;            // per-job context (e.g. sensor handle)
    TickType_t period;    // ticks between runs
    TickType_t next_run;  // tick of the next scheduled run
} job_t;

static esp_websocket_client_handle_t s_ws;
static job_t s_jobs[MAX_JOBS];
static int s_job_count;

static void job_add(job_fn_t fn, void *arg, TickType_t period)
{
    if (s_job_count < MAX_JOBS) {
        s_jobs[s_job_count++] = (job_t){ .fn = fn, .arg = arg,
                                          .period = period, .next_run = 0 };
    }
}

// One adapter per configured sensor; it keeps CONFIG_* ids and sensor
// details out of the job loop. register_jobs() is the only place that
// decides which sensors exist, so app_main stays free of #ifdefs.

#if CONFIG_DS18B20_GPIO != 0
static void ds18b20_job(void *arg)
{
    send_dsb_data(s_ws, CONFIG_DS18B20_SYSTEM_ID, arg);
}
#endif

#if CONFIG_SHT_ENCLOSURE_ID != 0
static void sht_job(void *arg) {
  send_sht_data(s_ws, CONFIG_SHT_ENCLOSURE_ID, arg);
}
#endif

#if CONFIG_BH_PLANT_SITE_ID != 0
static void bh_job(void *arg) {
    send_bh1750_data(s_ws, CONFIG_BH_PLANT_SITE_ID);
}
#endif

static void register_jobs(void) {
#if CONFIG_I2C_DATA_GPIO != 0
    ESP_ERROR_CHECK(i2cdev_init());
#endif
#if CONFIG_DS18B20_GPIO != 0
    onewire_bus_handle_t bus = get_bus_handle();
    ds18b20_device_handle_t device = get_ds18b20_handle(bus);
    dsb_handle_t dsb = { bus, device };
    job_add(ds18b20_job, &dsb, pdMS_TO_TICKS(SENSOR_PERIOD_MS));
#endif
#if CONFIG_SHT_ENCLOSURE_ID != 0
    sht4x_t sht4x_dev =  get_shtx_dev();

    job_add(sht_job, &sht4x_dev, pdMS_TO_TICKS(SENSOR_PERIOD_MS));
#endif
#if CONFIG_BH_PLANT_SITE_ID != 0
    get_bh1750_dev();

    job_add(bh_job, NULL, pdMS_TO_TICKS(SENSOR_PERIOD_MS));
#endif
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
    register_jobs();

    for (;;) {
        TickType_t now = xTaskGetTickCount();
        TickType_t sleep = pdMS_TO_TICKS(SENSOR_PERIOD_MS);

        for (int i = 0; i < s_job_count; i++) {
            job_t *job = &s_jobs[i];

            job->fn(job->arg);
        }

        vTaskDelay(sleep);
    }
}
