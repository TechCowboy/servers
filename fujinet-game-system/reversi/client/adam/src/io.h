#ifndef IO_H
#define IO_H

// This allows me to run this on an emulator
//#define NO_FUJI

#ifdef NO_FUJI

#define EOS_WRITE_CHARACTER_DEVICE FAKE_eos_write_character_device
#define EOS_READ_CHARACTER_DEVICE FAKE_eos_read_character_device
unsigned char FAKE_eos_write_character_device(unsigned char dev, void *buf, unsigned short len);
unsigned char FAKE_eos_read_character_device(unsigned char dev, void *buf, unsigned short len);

#else

#define EOS_WRITE_CHARACTER_DEVICE eos_write_character_device
#define EOS_READ_CHARACTER_DEVICE eos_read_character_device

#endif

#define NET_DEV 0x09
#define FUJI_DEV 0x0F
#define CHANNEL_MODE_JSON 0x01
#define ACK 0x80

#define DISPLAY_DEBUG 1
#define DEBUG_DELAY (400)

#define MAX_APP_DATA (128)
#define MAX_URL (256)
#define MAX_QUERY (128)

extern unsigned char response[1024];

typedef struct
{
    unsigned char century, // Century
        year,              // Year
        month,             // Month
        day,               // Day
        hour,              // Hour
        minute,            // Minute
        second;            // Second
} FUJI_TIME;

typedef struct
{
    unsigned char cmd;
    unsigned short mode;
    unsigned char trans;
    unsigned char url[MAX_URL];
} FUJI_CMD;

typedef struct
{
    unsigned char cmd;
    char mode
} FUJI_SET_CHANNEL;

typedef struct
{
    unsigned char cmd;
    char query[MAX_QUERY];
} FUJI_JSON_QUERY;

typedef struct
{
    unsigned char cmd;
    unsigned short creator;
    unsigned char app;
    unsigned char key;
} FUJI_APP;

typedef struct
{
    unsigned char cmd;
    unsigned short creator;
    unsigned char app;
    unsigned char key;
    char data[MAX_APP_DATA];
} FUJI_APP_DATA;

char *strncpy2(char *dest, char *src, size_t size);

/*
io_time
- This function gets the time from the fujinet device
Parameters
wait_until: FUJI_TIME representing the time to wait for.

Returns
    0: Success
    1: Could not open website
    3: Could not get result
*/
int io_time(FUJI_TIME *time);

void add_time(FUJI_TIME *result, FUJI_TIME *time1, FUJI_TIME *add_time);

bool time_reached(FUJI_TIME *wait_until);

    /*
    io_json_open
    - This function will open the website, switch to json mode and prepare for parsing
      After calling this function you may use io_json_query

    Parameters
      url: website to request the json information

    Returns
        0: Success
        1: Could not open website
        2: Could not switch to json mode
        3: Could not set parsing mode
    */
    int io_json_open(char *url);

/*
io_json_query
- After performing a io_json_open, request the json element

Parameters
- element:          the json element to collect
- data:             the data within the element as a string
                    if unsuccessful, data buffer will be an empty string
- max_buffer_size:  the maximum size of the string including the null terminator

Returns
    0: Success
    1: Did not receive acknowledgment of the query command
    2: Could not find the element

*/
int io_json_query(char *element, char *data, int max_buffer_size);

/*
io_json_close
- Close the website used by the last io_json_open.  No check is performed.

Returns
    0: Success

*/
int io_json_close(void);
#endif

