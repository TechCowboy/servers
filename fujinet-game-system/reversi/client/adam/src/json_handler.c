#include <stdbool.h>
#include <eos.h>
#include "sound.h"
#include <string.h>
#include <conio.h>
#include "io.h"
#include "json_handler.h"
#include"sound.h"

//#define NET_DIAGS

unsigned char response[1024];


static char _url[256];
static char _pieces[BOARD_SIZE*BOARD_SIZE+1];
static char _last_result[MAX_RESULT];
static char _last_data[256];
static char _table[MAX_TABLE_SIZE];
static char _my_name[MAX_NAME_SIZE];
static bool _connected = false;
static int  _turn = -1;
static char _players[2][MAX_NAME_SIZE];
static char _player_colors[2];
static int  _active_player = -1;
static int  _move_time = 0;
static int  _valid_moves[BOARD_SIZE*BOARD_SIZE];

static char _tables[MAX_TABLES][MAX_TABLE_SIZE];
static char _tables_desc[MAX_TABLES][MAX_TABLE_DESC_SIZE];

void debug_start(void);
void debug_end(void);

int reversi_init(char *url_in)
{
    int i;

    snprintf(_url, sizeof(_url), "N:%s", url_in);

    for(i=0;i<BOARD_SIZE*BOARD_SIZE;i++)
    {
        _pieces[i]=EMPTY;
    }

    for (i = 0; i < BOARD_SIZE * BOARD_SIZE; i++)
    {
        _valid_moves[i] = NULL;
    }

    for (i = 0; i < 2; i++)
    {
#ifdef NO_FUJI
        snprintf(_players[i], MAX_NAME_SIZE, "NOFUJI%d", i+1);
#else
        snprintf(_players[i], MAX_NAME_SIZE, "FUJI%d", i + 1);
#endif
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


bool is_connected(void)
{
    return _connected;
}

bool update_data(void)
{
    static bool first_time = true;
    bool data_change = false;
    int i, j, x, y, r;
    char query[MAX_QUERY_SIZE];


    _connected = true;
    do
    {

        // board
        if ((r = io_json_query("/bd", _pieces, sizeof(_pieces))) != 0)
        {
#ifdef NET_DIAGS
            sound_negative_beep();
            debug_start();
            cprintf("%s\n", query);
            debug_end();
#endif
            break;
        }

#ifdef NO_FUJI
        strcpy(_pieces, ".................W.WWW....WBB.....BBB...........................");
#endif

        // turn
        if ((r = io_json_query("/t", response, sizeof(response))) != 0)
        {
#ifdef NET_DIAGS
            sound_negative_beep();
            debug_start();
            cprintf("%s\n", query);
            debug_end();
#endif
            break;
        }
#ifdef NO_FUJI
        if (_move_time < 2)
            _turn += 1;
#else
        _turn = atoi(response);
#endif
        // active player
        if ((r = io_json_query("/a", response, sizeof(response))) != 0)
        {
#ifdef NET_DIAGS
            sound_negative_beep();
            debug_start();
            cprintf("%s\n", query);
            debug_end();
#endif
            break;
        }
#ifdef NO_FUJI
        if (_active_player)
            _active_player = 0;
        else
            _active_player = 1;
#else
        _active_player = atoi(response);
#endif

        // last result
        // if ((r = io_json_query("/l", _last_result, sizeof(_last_result))) != 0)
        //{
        //    sound_negative_beep();
        //    break;
        //}

        // move time left
        if ((r = io_json_query("/m", response, sizeof(response))) != 0)
        {
#ifdef NET_DIAGS
            sound_failure();
            debug_start();
            cprintf("%s\n", query);
            debug_end();
#endif
            break;
        }

#ifdef NO_FUJI
        if (_move_time < 0)
            _move_time = 30;
        else
            _move_time--;
#else
        _move_time = atoi(response);
#endif
        for (i = 0; i < 64; i++)
        {
            _valid_moves[i] = -1;

            snprintf(query, sizeof(query), "/vm/%d/m", i);
            // valid moves
            if ((r = io_json_query(query, response, sizeof(response))) != 0)
            {
                break;
            }

            _valid_moves[i] = atoi(response);
        }

        for (i = 0; i < 2; i++)
        {
            snprintf(query, sizeof(query), "/pl/%d/n", i);

            // players
            if ((r = io_json_query(query, response, sizeof(response))) != 0)
            {
#ifdef NET_DIAGS
                sound_negative_beep();
                debug_start();
                cprintf("%s\n", query);
                debug_end();
#endif
            }
            else
            {
#ifdef NO_FUJI
                snprintf(response, sizeof(response), "Player%d", i);
#endif

                strncpy(_players[i], response, MAX_NAME_SIZE);
            }

            snprintf(query, sizeof(query), "/pl/%d/c", i);
            // players
            if ((r = io_json_query(query, response, sizeof(response))) != 0)
            {
#ifdef NET_DIAGS
                sound_negative_beep();
                debug_start();
                cprintf("%s\n", query);
                debug_end();
#endif
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

        data_change = (!(_last_data == _pieces)) || first_time;

        first_time = false;
        _connected = true;
    } while (false);


    return data_change;
}

bool refresh_data(void)
{
    bool data_change = false;
    int i;
    
    snprintf(response, sizeof(response), "%s/state?player=%s&table=%s", _url, _my_name, _table);
    for (i = 0; i < sizeof(_pieces); i++)
        _pieces[i] = '\0';

    if (io_json_open(response) == 0)
    {
        data_change = update_data();
        io_json_close();
    }


    return data_change;
}

int get_turn(void)
{
    return _turn;
}

bool is_my_turn(int my_number)
{
    if (_active_player == my_number)
        return true;
    else
        return false;
}

void get_board(char *board_out, int size)
{
    if (size >= 64)
    {
        memcpy(board_out, _pieces, 64);
    } else
    {
        sound_negative_beep();
    }

}

void get_player_name(int num, char *player)
{
    if ((num >= 0) && (num < 2))
        strcpy(player, _players[num]);
}

char get_player_color(int num)
{
    if ((num >= 0) && (num < 2))
        return _player_colors[num];
    else
        return '?';
}

int get_active_player(void)
{
    return _active_player;
}

int put_move(int column, int row, char color)
{
    snprintf(response, sizeof(response), "%s/move/:\"%d\"?player=%s&table=%s", _url, row*BOARD_SIZE+column, _my_name, _table);
    if (io_json_open(response) == 0)
    {
        update_data();
        io_json_close();
        return 0;
    }
    return 1;
}

bool is_valid_move_debug(int column, int row)
{
    int position = row * BOARD_SIZE + column;
    bool valid = false;
    int i=0;

    debug_start();

    cprintf("column:%d row:%d looking for %d\n", column, row, position);

    while(_valid_moves[i] != -1)
    {
        cprintf("looking at %d\n", _valid_moves[i]);
        if (position == _valid_moves[i])
            break;
        i++;
    }

    cprintf("returning %d\n", _valid_moves[i]);

    debug_end();
    return _valid_moves[i] != -1;
}

bool is_valid_move(int column, int row)
{
    int position = row * BOARD_SIZE + column;
    bool valid = false;
    int i = 0;

    while (_valid_moves[i] != -1)
    {
        if (position == _valid_moves[i])
            break;
        i++;
    }

    return _valid_moves[i] != -1;
    //is_valid_move_debug(column, row);
}

int get_remaining_time(void)
{
    return _move_time;
}

void get_tables()
{
    int i, r;
    char query[MAX_QUERY_SIZE];

    for (i=0; i<MAX_TABLES; i++)
    {
        strcpy(_tables[i], "");
        strcpy(_tables_desc[i], "");
    }

    _connected = false;
    snprintf(response, sizeof(response), "%s/tables");
    if (io_json_open(response) == 0)
    {
        for (i=0; i<MAX_TABLES; i++)
        {
            snprintf(query, sizeof(query), "/t/%d", i);

            // table
            if ((r = io_json_query(query, response, sizeof(response))) != 0)
            {
#ifdef NET_DIAGS
                sound_negative_beep();
                debug_start();
                cprintf("%s\n", query);
                debug_end();
#endif
            }
            else
            {
#ifdef NO_FUJI
                snprintf(response, sizeof(response), "table%d", i);
#endif

                strncpy(_tables[i], response, MAX_TABLE_SIZE);
                _connected = true;
            }

            snprintf(query, sizeof(query), "/n/%d", i);

            // players
            if ((r = io_json_query(query, response, sizeof(response))) != 0)
            {
#ifdef NET_DIAGS
                sound_negative_beep();
                debug_start();
                cprintf("%s\n", query);
                debug_end();
#endif
            }
            else
            {
#ifdef NO_FUJI
                snprintf(response, sizeof(response), "table_desc%d", i);
#endif

                strncpy(_tables_desc[i], response, MAX_TABLE_DESC_SIZE);
            }
            return;
        }
        io_json_close();
        return;
    }
}

void get_table(int num, char *table, char *table_desc)
{
    strcpy(table, _tables[num]);
    strcpy(table_desc, _tables_desc[num]);
}
