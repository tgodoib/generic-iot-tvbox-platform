//
// Created by Tiago Bannwart on 16/06/26.
//

#ifndef SSAIOT_ESP_DATASOURCE_H
#define SSAIOT_ESP_DATASOURCE_H


class DataSource {
private:
    static unsigned int last_ok_read;
    static unsigned int max_down_time_s;
public:
    static void setup();
    static bool is_updated();
};


#endif //SSAIOT_ESP_DATASOURCE_H
