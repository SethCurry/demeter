#ifndef DSB_H_
#define DSB_H_

// DS18B20 water-temperature probe, exposed as a Demeter sensor.
//
// ds18b20_sensor() is what main registers; init/measure are public so the
// sensor can also be driven directly.

#include "demeter.h"

// Descriptor for the configured probe. The 1-Wire bus is created and the device
// is discovered by the descriptor's init hook.
demeter_sensor_t ds18b20_sensor(void);

esp_err_t ds18b20_sensor_init(void);
size_t ds18b20_sensor_measure(void *ctx,
                              demeter_measurement_t *measurements,
                              size_t max_measurements);

#endif
