#ifndef DEMETERWS_H_
#define DEMETERWS_H_

// == function prototypes =======================================
//
#include "esp_websocket_client.h"

void send_sensor_data(esp_websocket_client_handle_t client, int target_type, int target_id, int measurement);
esp_websocket_client_handle_t connect_websocket();

#endif
