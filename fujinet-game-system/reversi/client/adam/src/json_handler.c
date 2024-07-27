#include <stdbool.h>
#include <eos.h>
#include "sound.h"
#include <string.h>
#include <conio.h>
#include "io.h"
#include "json_handler.h"
#include"sound.h"


unsigned char response[1024];

static char _url[256];
static char _board[65];
static char _last_result[256];
static char _last_data[256];
static char _table[128];
static char _my_name[128];
static bool _connected = false;
static int  _turn = -1;
static char _players[2][128];
static char _player_colors[2];
static int  _active_player = -1;
static int  _move_time = 0;
static int  _valid_moves[64];




int reversi_init(char *url_in)
{
    int i;

    snprintf(_url, sizeof(_url), "N:%s", url_in);

    for(i=0;i<64;i++)
    {
        _board[i]=EMPTY;
    }

    for (i = 0; i < 64; i++)
    {
        _valid_moves[i] = NULL;
    }

    for (i = 0; i < 2; i++)
    {
        strcpy(_players[i], "");
    }
}

void set_table(char *table_in)
{
    strncpy2(_table, table_in, sizeof(_table));
}

void set_name(char *name)
{
    strncpy2(_my_name, name, sizeof(_my_name));
}

void set_num_players(int count)
{
    snprintf(response, sizeof(response), "%s/state/?count=%d&table=%s",_url, count, _table);
    if (io_json_open(response) == 0)
    {
        io_json_close();
    }
}





bool refresh_data(void)
{
static bool first_time = true;
       bool data_change = false;
       int i, j, x, y, r;
       char query[128];

    snprintf(response, sizeof(response), "%s/state?player=%s&table=%s", _url, _my_name, _table);
    for (i = 0; i < sizeof(_board); i++)
        _board[i] = '\0';

    if (io_json_open(response) == 0)
    {
        _connected = false;
        do {

            // board
            if ((r = io_json_query("/bd", _board, sizeof(_board))) != 0)
            {
                sound_negative_beep();
                break;
            }

#ifdef NO_FUJI
            strcpy(_board, "");
#endif


            // turn
            if ((r = io_json_query("/t", response, sizeof(response))) != 0)
            {
                sound_negative_beep();
                break;
            }
#ifdef NO_FUJI
            _turn += 1;
#else
            _turn = atoi(response);
#endif
            // active player
            if ((r = io_json_query("/a", response, sizeof(response))) != 0)
            {
                sound_negative_beep();
                break;
            }
#ifdef NO_FUJI
            strcpy(response, "0");
#endif
            _active_player = atoi(response);

            // last result
            //if ((r = io_json_query("/l", _last_result, sizeof(_last_result))) != 0)
            //{
            //    sound_negative_beep();
            //    break;
            //}

            // move time left
            if ((r = io_json_query("/m", response, sizeof(response))) != 0)
            {
                sound_negative_beep();
                break;
            }
            _move_time = atoi(response);

            for (i = 0; i < 64; i++)
            {
                _valid_moves[i] = -1;

                sprintf(query, "/vm/%d/m", i);
                // valid moves
                if ((r = io_json_query(query, response, sizeof(response))) != 0)
                {
                    break;
                }
            
                _valid_moves[i] = atoi(response);
            }

            for (i=0; i<2; i++)
            {
                sprintf(query, "/pl/%d/n", i);
                // players
                if ((r = io_json_query(query, response, sizeof(response))) != 0)
                {
                    sound_negative_beep();
                } else
                {
#ifdef NO_FUJI
                    sprintf(response, "Player%d", i);
#endif

                    strcpy(_players[i], response);
                }
                sprintf(query, "/pl/%d/c", i);
                // players
                if ((r = io_json_query(query, response, sizeof(response))) != 0)
                {
                    sound_negative_beep();
                }
                else
                {
#ifdef NO_FUJI
                    if (i == 0)
                        strcpy(response, "B");
                    else
                        strcpy(response, "W");
#endif
                    _player_colors[i] = response[0];
                }
            }



            data_change = (!(_last_data == _board)) || first_time;

            first_time = false;
            _connected = true;
        } while (false);
    }


    return data_change;
}

int get_turn(void)
{
    return _turn;
}

void get_board(char *board_out, int size)
{
    if (size >= 64)
    {
        memcpy(board_out, _board, 64);
    } else
    {
        sound_negative_beep();
    }

}

void get_player_name(int num, char *player)
{
    if ((num > 0) && (num < 2))
        strcpy(player, _players[num]);
}

char get_player_color(int num)
{
    if ((num > 0) && (num < 2))
        return _player_colors[num];
    else
        return '?';
}

int get_active_player(void)
{
    return _active_player;
}
