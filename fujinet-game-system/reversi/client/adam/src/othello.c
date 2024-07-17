
#include <msx.h>
#include <graphics.h>
#include <games.h>
#include "board.h"
#include "charset.h"
#include "spriteset.h"
#include "joystick.h"
#include "sound.h"
#include <conio.h>

#include <smartkeys.h>
#include <eos.h>
#include "io.h"
#include "json_handler.h"


/*
DefPatternTable EQU     0h
DefNameTable    EQU     1800h
DefSprAttrTable EQU     1b00h
DefColorTable   EQU     2000h
DefSprPatTable  EQU     3800h
*/

	// graphics mode 1
	// 32 x 24

unsigned char board[768];





#define LINE_START  7
#define LINE_STOP  17
#define LINE_WIDTH  6

int white_line = LINE_START;
int black_line = LINE_START;
#define WHITE_START_X  0
#define BLACK_START_X  25


#define CURSOR_TOP_LEFT      0
#define CURSOR_MIDDLE_LEFT   1
#define CURSOR_BOTTOM_LEFT   2

#define CURSOR_TOP_MIDDLE    3
#define CURSOR_MIDDLE_MIDDLE 4
#define CURSOR_BOTTOM_MIDDLE 5

#define CURSOR_TOP_RIGHT     6
#define CURSOR_MIDDLE_RIGHT  7
#define CURSOR_BOTTOM_RIGHT  8

int status_y = 22;

int line = 0;

int trace_on = 1;
int trig=1;

int print_trace(char *message);


char my_name[16] = { "TechCowboy" };
char their_name[16] = { "THEM" };
int game_type;


/* z88dk specific opt */
#pragma printf = "%c %u"
#ifdef SCCZ80
	void prtbrd(char b[64], bool mefirst) __z88dk_fastcall;
int prtscr(char b[64]) __z88dk_fastcall;
#endif




#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <stdlib.h>
#include <time.h>   /* Needed just for srand seed */


#ifndef fputc_cons
#define fputc_cons putchar
#endif

#ifndef getk
#define getk getchar
#endif

int handicap;
char selfplay;		/* true if computer playing with itself */
/* int h[4][2];	*/	/* handicap position table */
int h[8];		/* handicap position table */
char mine, his;		/* who has black (*) and white (@) in current game */
char mefirst;		/* true if computer goes first in current game */

struct mt {
		int x;
		int y;
		int c;
		int s;
	 };


int print(int x, int y, char *message, bool blit)
{
	int pos = y * 32 + x;
	unsigned int addr;

	for (x = 0; x < strlen(message); x++)
		board[pos + x] = message[x];

	if (blit)
	{
		addr = NameTable;
		msx_vwrite(board, addr, 768);
	}
	return 0;
}

int print_down(int x, int y, char *message, bool blit)
{
	int pos = y * 32 + x;
	unsigned int addr;

	for (x = 0; x < strlen(message); x++)
	{
		if (y > 20)
			break;

		board[pos] = message[x];
		pos += 32;
		y++;

	}

	if (blit)
	{
		addr = NameTable;
		msx_vwrite(board, addr, 768);
	}
	return 0;
}

int print_trace(char *message)
{
	int x,y, pos;

	if (trace_on)
	{
		line++;
		if (line >22)
		{
			line = 0;
			for (y = 0; y < 22 + 1; y++)
			{
				for (x = 0; x < WHITE_START_X + LINE_WIDTH; x++)
				{
					pos = y * 32 + x;
					board[pos] = ' ';
				}
			}
		}

		print(0, line, message, true);
	}
	return 0;
}

int print_error(char *message)
{
	int x, y, pos;

	for (y = 21; y < 24; y++) for (x = 0; x < 32; x++)
	{
		pos = y * 32 + x;
		board[pos] = ' ';
	}
	x = strlen(message) / 2 + 16;

	print(x, y, message, true);
	


	return 0;
}

int print_info(char *message)
{
	int x, y, pos;

	for (y = 21; y < 24; y++) for (x = 0; x < 32; x++)
	{
		pos = y * 32 + x;
		board[pos] = ' ';
	}
	x = 16 - strlen(message) / 2;
	if (x < 0)
		x = 0;
	y = status_y;

	print(x, y, message, true);

	return 0;

}

int print_no_clear(int y, char *message)
{
	int x;

	x = 16 - strlen(message) / 2;
	if (x < 0)
		x = 0;

	print(x, y, message, true);

	return 0;
}

int print_white_line(char *message)
{
	int x,y,pos;
	char c;
	if (white_line > LINE_STOP)
	{
		for(y=LINE_START; y<LINE_STOP+1; y++)
		{
			for (x = WHITE_START_X; x < WHITE_START_X + LINE_WIDTH; x++)
			{
				pos = y * 32 + x;
				board[pos] = WHITE_LINE_COLOUR;
			}
		}
		white_line = LINE_START;
	}

	for (x = 0; x < strlen(message); x++)
	{
		c = message[x];
		if (c >= '1' && c <= '8')
		{
			c = c - '0';
			c = white_numbers[c];
		}
		else if (c == '-')
			c = white_numbers[0];
		message[x] = c;
	}

	print(WHITE_START_X +(LINE_WIDTH/2)-strlen(message)/2, white_line, message, false);
	white_line++;

	return 0;
}

int print_black_line(char *message)
{
	int x, y, pos;
	char c;
	if (black_line > LINE_STOP)
	{
		for (y = LINE_START; y < LINE_STOP + 1; y++)
		{
			for (x = BLACK_START_X; x < BLACK_START_X + LINE_WIDTH; x++)
			{
				pos = y * 32 + x;
				board[pos] = BLACK_LINE_COLOUR;
			}
		}
		black_line = LINE_START;
	}

	for(x=0; x<strlen(message); x++)
	{
		c = message[x];
		if (c>='1' && c<='8')
		{
			c = c - '0';
			c = black_numbers[c];
		} else
			if (c == '-')
				c = black_numbers[0];
		message[x] = c;
	}

	print(BLACK_START_X + (LINE_WIDTH / 2) - strlen(message) / 2, black_line, message, false);
	black_line++;

	return 0;
}

void init_msx_graphics()
{
	init_character_set();

	init_adam_sprites();

	vdp_color(BACKGROUND_COLOUR_GRAPHICS);
}

void newbrd()
{
	memcpy(board, newboard, sizeof(newboard));
}

int cntbrd(char *b, char color)
{
	int count = 0;
	int i;

	for(i=0; i<64; i++)
		if (b[i] == color)
			count++;
	return count;
}

void prtbrd(char *b, bool mefirst)
{
	unsigned int addr;
	int x,y, pos;
	char black_count[3], 
	     white_count[3];

	newbrd();

	if (mefirst)
	{
		print_down(0, LINE_START + 1, their_name, false);
		print_down(31, LINE_START + 1, my_name, false);
	}
	else
	{
		print_down(0, LINE_START + 1, my_name, false);
		print_down(31, LINE_START + 1, their_name, false);
	}

	snprintf(black_count, sizeof(black_count), "%2d", cntbrd(b, BLACK));
	snprintf(white_count, sizeof(white_count), "%2d", cntbrd(b, WHITE));

	print(WHITE_START_X + 2, LINE_STOP+1, white_count, false);
	print(BLACK_START_X + 2, LINE_STOP+1, black_count, false);

	for (x = 0; x < 8; x++)
	{
		for(y=0; y<8; y++)
		{
			pos = y*64+x*2 + BOARD_START_Y*32 + BOARD_START_X;

			switch(b[y*8+x])
			{
				case BLACK:
					board[pos] 		= (unsigned char) (BLACK_TOP_LEFT     & 0xFF);
					board[pos + 1]  = (unsigned char) (BLACK_TOP_RIGHT    & 0xFF);
					board[pos + 32] = (unsigned char)(BLACK_BOTTOM_LEFT   & 0xFF);
					board[pos + 33] = (unsigned char)(BLACK_BOTTOM_RIGHT  & 0xFF);
					break;
				case WHITE:
					board[pos]      = (unsigned char)(WHITE_TOP_LEFT      & 0xFF);
					board[pos + 1]  = (unsigned char)(WHITE_TOP_RIGHT     & 0xFF);
					board[pos + 32] = (unsigned char)(WHITE_BOTTOM_LEFT   & 0xFF);
					board[pos + 33] = (unsigned char)(WHITE_BOTTOM_RIGHT  & 0xFF);
					break;
				default:
					break;
			}
		}
	}

	addr = NameTable;
	msx_vwrite(board, addr, 768);
}

int main()
{
	char b[64];
	int i;
	char key;
	char message[32];
	int waiting;
	int connection = 0;
	int last_error, bytes_waiting;
	char waitingstr[] = {"|/-\\"};
	bool game_in_progress = true;
	char url[256];
	FUJI_TIME adjust, future_time;

	memset(&adjust, 0, sizeof(FUJI_TIME));

	adjust.second = 10;

	for(i=0; i<sizeof(b); i++)
		b[i] = EMPTY;
	
	h[0] = h[1] = h[4] = h[7] = 0;
	h[2] = h[3] = h[5] = h[6] = 7;

	sound_init();
	smartkeys_set_mode();

	vdp_color(BACKGROUND_COLOUR_TEXT);

	vdp_set_mode(mode_2);

	

	cprintf(

		"#FUJINET REVERSI - ADAM EDITION\n"
		"      Adapted for Fujinet\n"
		"        by Norman Davie\n\n"
		" Reversi is a strategy game\n"
		"invented during the Victorian\n"
		"era.  The goal: to have the\n"
		"majority of pieces on the board\n"
		"your own colour.  Trapping\n"
		"opposing pieces between your\n"
		"coloured pieces converts them\n"
		"to your colour."
		"\n\n"
		"BLACK is always the FIRST\n"
		"player\n");


	sound_mode_change();

	smartkeys_display("1 Player\n  Local", "2 Player\n  Local", "  Host\n  Game", " Remote\n  Host", NULL, "  QUIT");

	smartkeys_display(NULL, NULL, NULL, NULL, NULL, NULL);


	//clrscr();

	sound_chime();

	vdp_color(BACKGROUND_COLOUR_GRAPHICS);
	vdp_set_mode(mode_2);

	strncpy2(url, "http://192.168.2.184:8080", sizeof(url));

		do 
	{
		sound_chime();

		init_msx_graphics();
		newbrd();
		
		reversi_init(url);
		set_name(my_name);
		set_table("bot2a");

		prtbrd(b, mefirst);
		io_time(&future_time);
		while (game_in_progress)
		{
			if (time_reached(&future_time))
			{
				sound_chime();
				add_time(&future_time, &future_time, &adjust);
				refresh_data();
				get_board(b, sizeof(b));

				game_in_progress = get_round() != 5;
				
				prtbrd(b, mefirst);
			}
		}

	} while (game_in_progress);

	return 0;
}


