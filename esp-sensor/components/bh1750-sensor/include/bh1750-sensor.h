#ifndef BH1750_SENSOR_H_
#define BH1750_SENSOR_H_

// BH1750 ambient-light sensor for a plant site, exposed as a Demeter sensor.

#include "demeter.h"

demeter_sensor_t bh1750_sensor(void);

esp_err_t bh1750_sensor_init(void);
size_t bh1750_sensor_measure(void *ctx,
                             demeter_measurement_t *measurements,
                             size_t max_measurements);

#endif
