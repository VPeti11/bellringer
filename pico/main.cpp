#include "pico/stdlib.h"
#include <string>

int main() {
    stdio_init_all();

    const uint LED_PIN = 1;
    gpio_init(LED_PIN);
    gpio_set_dir(LED_PIN, GPIO_OUT);

    std::string buffer;

    while (true) {
        int ch = getchar_timeout_us(0);

        if (ch == PICO_ERROR_TIMEOUT) {
            continue;
        }

        if (ch == '\n' || ch == '\r') {
            if (buffer == "HIGH") {
                gpio_put(LED_PIN, 1);
                printf("OK HIGH\n");
            } else if (buffer == "LOW") {
                gpio_put(LED_PIN, 0);
                printf("OK LOW\n");
            } else if (!buffer.empty()) {
                printf("ERR UNKNOWN\n");
            }
            buffer.clear();
        } else {
            buffer.push_back((char)ch);
        }
    }
}
