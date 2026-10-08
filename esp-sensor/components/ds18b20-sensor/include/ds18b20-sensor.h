#ifndef DSB_H_
#define DSB_H_

// == function prototypes =======================================
//
#include "esp_websocket_client.h"
#include "ds18b20.h"


typedef struct {
    onewire_bus_handle_t bus;
    ds18b20_device_handle_t ds18b20;
} dsb_handle_t;

onewire_bus_handle_t get_bus_handle();
ds18b20_device_handle_t get_ds18b20_handle(onewire_bus_handle_t bus);
void send_dsb_data(esp_websocket_client_handle_t client, int target_id, dsb_handle_t *dsb);

#endif
