#ifndef SHT_SENSOR_H_
#define SHT_SENSOR_H_

#include "esp_websocket_client.h"
#include "bh1750.h"

void send_bh1750_data(esp_websocket_client_handle_t client, int target_id);

void get_bh1750_dev();

#endif
