idf_component_register(SRCS "esp-sensor.c" "wifi.c"
                    INCLUDE_DIRS "."
                    REQUIRES "esp_wifi" "freertos" "nvs_flash" "demeterws" "ds18b20-sensor" "sht4x-sensor"
                    PRIV_REQUIRES "esp_driver_gpio")
