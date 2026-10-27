#ifndef SHT_SENSOR_H_
#define SHT_SENSOR_H_

// SHT4x temperature/humidity sensor for the enclosure, exposed as a Demeter
// sensor. One read produces two measurements (temperature and humidity).

#include "demeter.h"

demeter_sensor_t sht4x_sensor(void);

esp_err_t sht4x_sensor_init(void);
size_t sht4x_sensor_measure(void *ctx,
                            demeter_measurement_t *measurements,
                            size_t max_measurements);

#endif
