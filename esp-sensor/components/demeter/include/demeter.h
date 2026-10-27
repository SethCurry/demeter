#ifndef DEMETER_H_
#define DEMETER_H_

// Demeter sensor abstraction.
//
// A sensor component describes itself with a demeter_sensor_t: an optional
// one-time init hook, a measure hook and the period at which it wants to be
// read. app_main hands those descriptors to an orchestrator_t, which owns the
// single loop that runs them. That keeps every blocking operation (1-wire
// conversions, camera captures, websocket sends) on one task instead of one
// task or timer callback per sensor.

#include <stddef.h>
#include <stdint.h>

#include "esp_err.h"
#include "esp_websocket_client.h"
#include "freertos/FreeRTOS.h"

#ifdef __cplusplus
extern "C" {
#endif

// Storage is static, so both limits are compile-time and small on purpose.
#define DEMETER_MAX_SENSORS 8
#define DEMETER_MAX_MEASUREMENTS 4

typedef struct {
    int target_type;   // measurement type understood by the server
    int target_id;     // device/site the measurement belongs to
    int measurement;   // value, scaled by the sensor component
} demeter_measurement_t;

// Runs once when the sensor is registered with an orchestrator. May be NULL.
typedef esp_err_t (*demeter_sensor_init_t)(void);

// Samples the sensor and writes up to max_measurements values. Returns how many
// values were written (0 when the read failed). `ctx` is the descriptor's ctx.
typedef size_t (*demeter_sensor_measure_t)(void *ctx,
                                           demeter_measurement_t *measurements,
                                           size_t max_measurements);

typedef struct {
    const char *name;               // used for logging
    demeter_sensor_init_t init;     // optional
    demeter_sensor_measure_t measure;
    void *ctx;                      // passed back to measure, may be NULL
    uint32_t interval_ms;           // desired period between measurements
} demeter_sensor_t;

typedef struct {
    demeter_sensor_t sensor;
    TickType_t next_run;            // tick at which the sensor is due again
} demeter_sensor_slot_t;

typedef struct {
    demeter_sensor_slot_t slots[DEMETER_MAX_SENSORS];
    size_t num_sensors;
} orchestrator_t;

esp_websocket_client_handle_t connect_websocket(void);
void send_sensor_data(esp_websocket_client_handle_t client, int target_type,
                      int target_id, int measurement);

// Measures one sensor now and publishes everything it returned. Returns the
// number of measurements published.
size_t measure_sensor(esp_websocket_client_handle_t client,
                      const demeter_sensor_t *sensor);

// Resets an orchestrator to "no sensors".
esp_err_t orchestrator_init(orchestrator_t *orchestrator);

// Takes a copy of the descriptor, runs its init hook and schedules it to run on
// the next pass. Returns the init hook's error when the sensor cannot start.
esp_err_t orchestrator_add_sensor(orchestrator_t *orchestrator,
                                  demeter_sensor_t sensor);

// Measures every sensor whose interval has elapsed, publishing the results.
void orchestrator_measure(orchestrator_t *orchestrator,
                          esp_websocket_client_handle_t client);

// The loop body: measure what is due, then sleep until the next sensor is due.
// Call this forever from a task (app_main already is one).
void orchestrator_run(orchestrator_t *orchestrator,
                      esp_websocket_client_handle_t client);

#ifdef __cplusplus
}
#endif

#endif
