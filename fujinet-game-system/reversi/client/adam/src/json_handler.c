#include <stdbool.h>
#include <eos.h>
#include <string.h>
#include <conio.h>
#include "io.h"
#include "json_handler.h"


unsigned char response[1024];

static char _url[256];
static char _board[64];
static char _last_result[256];
static char _last_data[256];
static char _table[128];
static char _my_name[128];
static bool _connected = false;
static int  _round = -1;
static char _players[256];
static int  _active_player = -1;
static int  _move_time = 0;
static char _valid_moves[512];








int reversi_init(char *url_in)
{
    int i;

    snprintf(_url, sizeof(_url), "N:%s", url_in);

    for(i=0;i<64;i++)
    {
        _board[i]=EMPTY;
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

    snprintf(response, sizeof(response), "%s/state?player=%s&table=%s", _url, _my_name, _table);
    if (io_json_open(response) == 0)
    {
        io_json_query("/bd", _board, sizeof(_board));
        io_json_query("/l",  _last_result, sizeof(_last_result));
        io_json_query("/r",  response, sizeof(response));
        _round = atoi(response);
        io_json_query("/a",  response, sizeof(response));
        _active_player = atoi(response);
        io_json_query("/m", response, sizeof(response));
        _move_time = atoi(response);
        io_json_query("/v", response, sizeof(response));
        io_json_query("/vm", _valid_moves, sizeof(_valid_moves));
        io_json_query("/pl", _players, sizeof(_players));
        
        data_change = (! (_last_data == _board)) || first_time;

        first_time = false;
        _connected = true;
    } else
        _connected = false;

    return data_change;
}

int get_round(void)
{
    return _round;
}
void get_board(char *board_out, int size)
{
    memcpy(board_out, _board, size);
}


/*
    def __init__(self,url):
        self.url = url
        self.json_data = None
        self.json_action_data = None
        self.data_change = False
        self.last_data = "@"
        self.table = ""
        self.board = ""
        self.connected = False
        self.my_name = ""

        self.key_last_result = 'l'
        self.key_players      = 'pl'
        self.key_valid_moves = 'vm'
        self.key_active_player = 'a'
        self.key_name = 'n'
        self.key_move = 'm'
        self.key_board = 'bd'

    def set_table(self, table):
        self.table = table

    def set_name(self, name):
        self.my_name = name

    def set_players(self, players):
        request = self.url+"/state?count="+str(players)+"&table="+self.table
        requests.get(request)

    def refresh_data(self):
        try:
            request = self.url+"/state?player="+self.my_name+"&table="+self.table
            response = requests.get(request)
            self.json_data = json.loads(response.text)
            self.data_change = not (self.last_data == response.text)
            self.last_data = response.text;
            if self.data_change:
                print(response.text)
            self.connected = True
        except Exception as e:
            print(f"error: {e}")
            self.connected = False
        return self.data_change

    def get_board(self):
        return self.json_data[self.key_board]

    def send_action(self, action):
        success = True
        try:
            request = self.url+"/move/"+action+"?player="+self.my_name+"&table="+self.table
            response = requests.get(request)
            print(f"***send_action: {action}")
            self.json_action_data = json.loads(response.text)
            self.connected = True
        except:
            success = False
            self.connected = False
        return success

    def get_number_of_players(self):
        num = len(self.json_data[self.key_players])
        return num

    def get_name(self,player_num):
        return self.json_data[self.key_players][player_num][self.key_name]

    def get_playing(self, player_num):
        player = self.json_data[self.key_active_player]
        return player

    def get_last_result(self):
        return self.json_data[self.key_last_result]

    def get_active_player(self):
        return self.json_data[self.key_active_player]

    def get_valid_buttons(self):
        valid_moves = None

        moves = {}
        i = 0
        no_error = True
        while no_error:
            try:
                move = self.json_data[self.key_valid_moves][i][self.key_move]
                name = self.json_data[self.key_valid_moves][i][self.key_name]
                moves[move] = name
                i += 1
            except:
                no_error=False

        return moves
*/
