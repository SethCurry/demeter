idf_component_register(SRCS "timelapse-sensor.c"
                    PRIV_INCLUDE_DIRS .
                    REQUIRES demeter
                    PRIV_REQUIRES nvs_flash esp_psram esp32-camera esp_wifi esp_websocket_client
                    INCLUDE_DIRS "include")
