idf_component_register(SRCS "sht4x-sensor.c"
		    REQUIRES "sht4x" "esp_websocket_client" "demeterws"
                    INCLUDE_DIRS "include")
