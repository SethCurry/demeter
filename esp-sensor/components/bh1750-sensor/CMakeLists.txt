idf_component_register(SRCS "bh1750-sensor.c"
           		    REQUIRES "bh1750" "esp_websocket_client" "demeter"
                    INCLUDE_DIRS "include")
