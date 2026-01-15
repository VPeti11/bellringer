cmake_minimum_required(VERSION 3.13)

include($ENV{PICO_SDK_PATH}/pico_sdk_init.cmake)

project(pico_gpio_serial)

pico_sdk_init()

add_executable(pico_gpio_serial
    main.cpp
)

target_link_libraries(pico_gpio_serial pico_stdlib)

pico_enable_stdio_usb(pico_gpio_serial 1)
pico_enable_stdio_uart(pico_gpio_serial 0)

pico_add_extra_outputs(pico_gpio_serial)
