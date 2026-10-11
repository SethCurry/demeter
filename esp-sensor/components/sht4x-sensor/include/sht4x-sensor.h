#ifndef SHT_SENSOR_H_
#define SHT_SENSOR_H_

#include "esp_websocket_client.h"
#include "sht4x.h"


sht4x_t get_shtx_dev();
void send_sht_data(esp_websocket_client_handle_t client, int target_id,
                   sht4x_t *dev);

#endif
