
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
char query[50] = {""};
char url[256];
char message[32];
char player1_color;
char player2_color;
char my_color;
char my_number;

#define LINE_START  7
#define LINE_STOP  17
#define LINE_WIDTH  6

int white_line = LINE_START;
int black_line = LINE_START;
#define WHITE_START_X  0
#define BLACK_START_X  25

#define TURN_X 27
#define TURN_Y 0

#define TIME_X 27
#define TIME_Y 1

#define REVERSI_X 12

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

int line = 2;

int trace_on = 1;
int trig=1;

int print_trace(char *message);


char my_name[16] = { "TechCowboy" };
char black_player[16];
char white_player[16];
int turn;
char b[64];

/* z88dk specific opt */
#pragma printf = "%c %u"
#ifdef SCCZ80
	void prtbrd(char *b, bool mefirst, int turn, int remaining_time) __z88dk_fastcall;
int prtscr(char *b) __z88dk_fastcall;
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
char mine, his;		/* who has black (*) and white (@) in current game */
char mefirst;		/* true if computer goes first in current game */
int ap;
int remaining_time;

struct mt {
		int x;
		int y;
		int c;
		int s;
	 };

void init_msx_graphics();

void debug_start(void)
{
	vdp_color(BACKGROUND_COLOUR_TEXT);

	vdp_set_mode(mode_2);
	clrscr();
	gotoxy(0,0);
}

void debug_end(void)
{
	char wait[10];
	gets(wait);

	init_msx_graphics();

	vdp_color(BACKGROUND_COLOUR_GRAPHICS);
	vdp_set_mode(mode_2);
	prtbrd(b, mefirst, turn, remaining_time);
}

int print(int x, int y, char *message, bool blit)
{
	int pos = y * 32 + x;
	unsigned int addr;

	for (x = 0; x < strlen(message); x++)
		board[pos + x] = message[x];

	if (blit)
	{
		addr = NameTable;
		msx_vwrite(board, addr, sizeof(board));
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
		msx_vwrite(board, addr, sizeof(board));
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

bool wait_for_fire(void)
{
	int trig;

	read_joystick(&trig);
	
	return trig == 0;
}


bool getmov_local(char *b, int *column, int *row, char my_color)
{
	int joy = 0;
	int trig = 1;
	//	int c;
	int x, y, i;
	char atari_output[32];
	bool triggered = false;
	bool moving=true;

	while(moving) 
	{
		for (i=0; i<4; i++)
		{
			movsprite(*column, *row, MOVING_COLOR);

			joy = read_joystick(&trig);

			if (joy & UP)
				(*row)--;
			if (joy & DOWN)
				(*row)++;
			if (joy & LEFT)
				(*column)--;
			if (joy & RIGHT)
				(*column)++;

			moving =  joy & (UP | DOWN | LEFT | RIGHT | (trig==0));

			if (*row < 0)
				*row = 7;

			if (*row > 7)
				*row = 0;

			if (*column < 0)
				*column = 7;

			if (*column > 7)
				*column = 0;

			movsprite(*column, *row, MOVING_COLOR);


			// SEND ACTION

			if (trig == 0)
			{
				sound_chime();
				sound_chime();
				triggered = true;

				movsprite(*column, *row, SELECTED_COLOR);

				if (is_valid_move(*column, *row))
				{
					triggered = true;
					moving = false;
					break;
				}
				else
				{
					print_info("*** Illegal Move ***");
					sound_negative_beep();
				}
			}
			csleep(4);
		}

	}

	return triggered;
}

void init_msx_graphics()
{
	init_character_set();

	init_adam_sprites();

	vdp_color(BACKGROUND_COLOUR_GRAPHICS);
}

void newbrd()
{
	int i;
	memcpy(board, newboard, sizeof(newboard));
	for(i=0; i<64; i++)
		b[i] = EMPTY;

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

void prtbrd(char *b, bool mefirst, int turn, int remaining_time)
{
	unsigned int addr;
	int x,y, pos;
	char turn_str[32];
	char reversi_top[8];
	char reversi_bottom[8];
	char black_count[3],
		white_count[3];

	if (mefirst && (ap == 0))
	{
		snprintf(reversi_top, sizeof(reversi_top), "%s", black_reversi);
		snprintf(reversi_bottom, sizeof(reversi_bottom), "%s", mirror_white_reversi);
	} else
	{
		snprintf(reversi_top, sizeof(reversi_top), "%s", white_reversi);
		snprintf(reversi_bottom, sizeof(reversi_bottom), "%s", mirror_black_reversi);
	}
	print(REVERSI_X, 0, reversi_top, false);
	print(REVERSI_X, 1, reversi_bottom, false);

	snprintf(turn_str, sizeof(turn_str), "%2d", turn);
	print(TURN_X, TURN_Y, turn_str, false);

	if (remaining_time < 0) 
		remaining_time = 0;

	snprintf(turn_str, sizeof(turn_str), "%2d", remaining_time);
	print(TIME_X, TIME_Y, turn_str, false);

	print_down(0,  LINE_START + 1, white_player, false);
	print_down(31, LINE_START + 1, black_player, false);

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
	msx_vwrite(board, addr, sizeof(board));

	
}

/*
void showTableSelectionScreen()
{
	static uint8_t shownChip;
	static unsigned char tableIndex = 0;
	static uint8_t skipApiCall;

	skipApiCall = 0;
	// An empty query means a table needs to be selected
	while (strlen(query) == 0)
	{

		if (!skipApiCall)
		{
			// Show the status immediately before retrival
			centerStatusText("REFRESHING TABLE LIST..");
			// drawStatusTimer();
			drawBuffer();
			resetScreenWithBorder();
		}

		centerText(3, "CHOOSE A TABLE TO JOIN");
		drawText(6, 6, "TABLE");
		drawText(WIDTH - 13, 6, "PLAYERS");
		drawLine(6, 7, WIDTH - 12);

		drawBuffer();
		waitvsync();

		if (skipApiCall || apiCall("tables"))
		{
			if (!skipApiCall)
			{
				updateState(true);
			}
			skipApiCall = 0;
			if (tableCount > 0)
			{
				for (i = 0; i < tableCount; ++i)
				{
					drawText(6, 8 + i * 2, state.tables[i].name);
					drawText((unsigned char)(WIDTH - 6 - strlen(state.tables[i].players)), 8 + i * 2, state.tables[i].players);
					if (state.tables[i].players[0] > '0')
					{
						drawText((unsigned char)(WIDTH - 6 - strlen(state.tables[i].players) - 2), 8 + i * 2, "*");
					}
				}
			}
			else
			{
				centerText(12, "SORRY, NO TABLES ARE AVAILABLE");
			}

			// drawStatusText(" R+EFRESH  H+ELP  C+OLOR  S+OUND  Q+UIT");
			drawStatusText("R-EFRESH   H-ELP  C-OLOR   N-AME   Q-UIT");
			drawBuffer();
			disableDoubleBuffer();
			shownChip = 0;

			clearCommonInput();
			while (!inputTrigger || !tableCount)
			{
				readCommonInput();

				if (inputKey == 'h' || inputKey == 'H')
				{
					showHelpScreen();
					break;
				}
				else if (inputKey == 'r' || inputKey == 'R')
				{
					break;
				}
				else if (inputKey == 'c' || inputKey == 'C')
				{
					prefs[PREF_COLOR] = cycleNextColor() + 1;
					savePrefs();
					enableDoubleBuffer();
					skipApiCall = 1;
					break;
				}
				else if (inputKey == 'n' || inputKey == 'N')
				{
					showPlayerNameScreen();
					break;
				}
				else if (inputKey == 'q' || inputKey == 'Q')
				{
					quit();
				} 

				if (!shownChip || (tableCount > 0 && inputDirY))
				{

					drawText(4, 8 + tableIndex * 2, " ");
					tableIndex += inputDirY;
					if (tableIndex == 255)
						tableIndex = tableCount - 1;
					else if (tableIndex >= tableCount)
						tableIndex = 0;

					drawChip(4, 8 + tableIndex * 2);

					soundCursor();
					shownChip = 1;
				}
			}

			enableDoubleBuffer();

			if (inputTrigger)
			{
				soundSelectMove();

				// Clear screen and write server name
				resetScreenWithBorder();
				clearStatusBar();
				centerText(15, state.tables[tableIndex].name);

				strcpy(query, "?table=");
				strcat(query, state.tables[tableIndex].table);
				strcpy(tempBuffer, serverEndpoint);
				strcat(tempBuffer, query);

				//  Update server app key in case of reboot
				write_appkey(AK_LOBBY_CREATOR_ID, AK_LOBBY_APP_ID, AK_LOBBY_KEY_SERVER, tempBuffer);
			}
		}
	}

	centerText(17, "CONNECTING TO SERVER");
	drawBuffer();

	progressAnim(19);

	tableActionJoinServer();
}
*/

void help(void)
{
	unsigned char key;
	bool stay = true;

	vdp_color(BACKGROUND_COLOUR_TEXT);

	vdp_set_mode(mode_2);

	gotoxy(0,0);
	cprintf(
		__DATE__ "  " __TIME__ "\n"
				 " #FUJINET GAME SERVER REVERSI \n"
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
				 "player\n"
				 "\n\n\n"
				 "1=RETURN");


	stay = true;
	while(stay)
	{
		key = eos_read_keyboard();

		switch (key)
		{
		case SMARTKEY_I:
			sound_confirm();
			stay=false;
			return;
		case SMARTKEY_II:
			sound_negative_beep();
			break;
		case SMARTKEY_III:
			sound_negative_beep();
			break;
		case SMARTKEY_IV:
			sound_negative_beep();
			break;
		case SMARTKEY_VI:
			sound_negative_beep();
			break;
		default:
			sound_negative_beep();
			break;
		}
	}
	init_msx_graphics();
	clrscr();
	newbrd();
	get_board(b, sizeof(b));
	prtbrd(b, mefirst, turn, 0);
}

void read_all_data(void)
{
	clrscr();
	vdp_color(BACKGROUND_COLOUR_TEXT);

	vdp_set_mode(mode_2);

	snprintf(message, sizeof(message), "Waiting for connection...");
	gotoxy(16 - strlen(message) / 2, 12);
	cprintf(message);
	while (!is_connected())
	{
		refresh_data();
		csleep(10);
	}
	clrscr();
	snprintf(message, sizeof(message), "Connected!");
	gotoxy(16 - strlen(message) / 2, 12);

	csleep(40);

	turn = get_turn();
	player1_color = get_player_color(0);
	player2_color = get_player_color(1);

	if (player1_color == BLACK)
	{
		get_player_name(0, black_player);
		get_player_name(1, white_player);
	}
	else
	{
		get_player_name(0, white_player);
		get_player_name(1, black_player);
	}

	if (stricmp(black_player, my_name) == 0)
	{
		my_color = player1_color;
		my_number = 0;
		mefirst = true;
	}
	else
	{
		my_color = player2_color;
		my_number = 1;
		mefirst = false;
	}

	init_msx_graphics();
	newbrd();

	get_board(b, sizeof(b));

	prtbrd(b, mefirst, turn, 0);

}

int main()
{
	int column,row;
	char key;
	int waiting;
	int connection = 0;
	int last_error, bytes_waiting;
	char waitingstr[] = {"|/-\\"};
	bool game_in_progress = true;
	FUJI_TIME adjust, future_time;
	int timer=0;
	bool send_move = false;
	int i;
	bool restart = false;
	bool leave_table = true;
	bool first_time_move = true;

	memset(&adjust, 0, sizeof(FUJI_TIME));

	adjust.second = 1;

	for(i=0; i<sizeof(b); i++)
		b[i] = EMPTY;

	sound_init();

	sound_mode_change();

	vdp_color(BACKGROUND_COLOUR_GRAPHICS);
	vdp_set_mode(mode_2);

	strncpy2(url, "http://192.168.2.184:8080", sizeof(url));

	reversi_init(url);
	set_name(my_name);

	set_table("bot1a");

	do 
	{

		read_all_data();

		io_time(&future_time);

		timer = 0;
		restart = false;
		game_in_progress = true;
		first_time_move = true;

		while (game_in_progress)
		{
			timer--;

			//key = eos_read_keyboard();
			key = 0;
			switch (key)
			{
			case SMARTKEY_I:
				sound_confirm();
				help();
				break;
			case SMARTKEY_II:
				sound_negative_beep();
				break;
			case SMARTKEY_III:
				sound_negative_beep();
				break;
			case SMARTKEY_IV:
				sound_negative_beep();
				break;
			case SMARTKEY_VI:
				sound_confirm();
				leave_table = true;
				break;
			default:
				break;
			}

			if (is_my_turn(my_number))
			{
				if (first_time_move)
				{
					get_board(b, sizeof(b));

					prtbrd(b, mefirst, turn, remaining_time);

					sound_chime();

					showsprite(true);
					first_time_move = false;
				}

				csleep(1);
				if (! send_move)
				{
					send_move = getmov_local(b, &column, &row, my_color);
				}

				if (send_move)
				{
					if (is_connected())
					{
						sound_confirm();

						showsprite(false);

						put_move(column, row, my_color);
						io_time(&future_time);
						timer = 0;
						first_time_move = true;
						send_move = false;
					} else
					{
						read_all_data();
					}
				}

			} else // end my turn
			{

				showsprite(false);
				first_time_move = true;
			}

			if (timer <= 0)
			{
				timer = 10;

				if (time_reached(&future_time))
				{
					refresh_data();

					// only update the time if we're connected
					if (is_connected())
					{
						add_time(&future_time, &future_time, &adjust);
					} else
					{
						sound_failure();
						read_all_data();
					}

					turn = get_turn();
					if (turn == 0)
					{
						newbrd();
					}

					get_board(b, sizeof(b));
					ap = get_active_player();

					remaining_time = get_remaining_time();

					prtbrd(b, mefirst, turn, remaining_time);

					// if we're not connected, then let the user know
					if (! is_connected())
					{
						read_all_data();

						ap = -1;
					}

					switch(ap)
					{
						case 0: 
							snprintf(message, sizeof(message), "%s'S TURN", black_player);
							break;
						case 1:
							snprintf(message, sizeof(message), "%s'S TURN", white_player);
							break;
						default:
							snprintf(message, sizeof(message), "<<COMMUNICATION ERROR>>");
							//restart = true;
							break;
					}

					print_info(message);

				} // time reached


			} // timer <= 0
			csleep(1);
			if (restart)
			{
				break;
			}
		} // while game in progress


	} while (leave_table);


	return 0;
}


