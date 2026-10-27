idf_component_register(SRCS "ds18b20-sensor.c"
                    REQUIRES "ds18b20" "demeter" "onewire_bus"
                    INCLUDE_DIRS "include")
