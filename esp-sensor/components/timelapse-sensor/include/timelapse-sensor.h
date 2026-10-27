#ifndef TIMELAPSE_SENSOR_H_
#define TIMELAPSE_SENSOR_H_

// Timelapse camera, exposed as a Demeter sensor. It has no scalar measurement:
// the measure hook streams the JPEG over the websocket client it was created
// with, so the orchestrator only has to decide when to run it.

#include "demeter.h"

#define CAM_PIN_PWDN -1
#define CAM_PIN_RESET -1   //software reset will be performed
#define CAM_PIN_VSYNC 6
#define CAM_PIN_HREF 7
#define CAM_PIN_PCLK 13
#define CAM_PIN_XCLK 15
#define CAM_PIN_SIOD 4
#define CAM_PIN_SIOC 5
#define CAM_PIN_D0 11
#define CAM_PIN_D1 9
#define CAM_PIN_D2 8
#define CAM_PIN_D3 10
#define CAM_PIN_D4 12
#define CAM_PIN_D5 18
#define CAM_PIN_D6 17
#define CAM_PIN_D7 16

// A 5 MP JPEG streamed in 512-byte chunks takes seconds, so the timelapse runs
// far less often than the sensors.
#define TIMELAPSE_SENSOR_INTERVAL_MS (5 * 60 * 1000)

// `client` is handed to the measure hook, which streams the picture over it.
demeter_sensor_t timelapse_sensor(esp_websocket_client_handle_t client);

esp_err_t timelapse_sensor_init(void);
size_t timelapse_sensor_measure(void *ctx,
                                demeter_measurement_t *measurements,
                                size_t max_measurements);

esp_err_t init_camera(void);

void maybe_init_autofocus(void);

void send_demeter_timelapse(esp_websocket_client_handle_t client);

#endif
