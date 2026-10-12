idf_component_register(SRCS "timelapse-sensor.c"
                    PRIV_INCLUDE_DIRS .
                    PRIV_REQUIRES nvs_flash esp_psram esp_camera esp_wifi demeterws esp_websocket_client
                    INCLUDE_DIRS "include")
