idf_component_register(SRCS "dht-sensor.c"
                    INCLUDE_DIRS "include"
                    REQUIRES "freertos" "demeterws"
                    PRIV_REQUIRES "esp_driver_gpio")
