#ifndef TIMELAPSE_SENSOR_H_
#define TIMELAPSE_SENSOR_H_

#include "esp_websocket_client.h"

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

void send_demeter_timelapse(esp_websocket_client_handle_t client);

esp_err_t init_camera(void);

void maybe_init_autofocus(void);


#endif
