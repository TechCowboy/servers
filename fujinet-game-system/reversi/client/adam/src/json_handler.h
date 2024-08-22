/**
 * Weather
 *
 * Based on @bocianu's code
 *
 * @author Thomas Cherryhomes
 * @email thom dot cherryhomes at gmail dot com
 *
 */

#ifndef JSON_HANDLER
#define JSON_HANDLER

#include <stdbool.h>
#include <stddef.h>

#define BLACK 'B'
#define WHITE 'W'
#define EMPTY '.'

#define BOARD_SIZE 8

#define MAX_TABLES 32
#define MAX_TABLE_SIZE 32
#define MAX_TABLE_DESC_SIZE 32
#define MAX_NAME_SIZE 32
#define MAX_RESULT 80
#define MAX_QUERY_SIZE 128


int reversi_init(char *url);
void set_table(char *table_in);
void set_name(char *name);
void set_num_players(int count);
bool refresh_data(void) ;
void get_board(char *board_out, int size);
 int get_turn(void);
void get_player_name(int num, char *player);
char get_player_color(int num);
 int get_active_player(void);
 int put_move(int column, int row, char color);
bool is_valid_move(int column, int row);
bool is_my_turn(int my_number);
 int get_remaining_time(void);
bool is_connected(void);
void get_tables();
void get_table(int num, char *table, char *table_desc);

#endif /* JSON_HANDLER */
