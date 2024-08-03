#include "spriteset.h"
#include <msx.h>
#include <conio.h>

SPRITE_ATTRIBUTE sprite_attrib[TOTAL_SPRITES];

void init_adam_sprites()
{
	int s;

	for (s = 0; s < TOTAL_SPRITES; s++)
	{
		sprite_attrib[s].x = 0;
		sprite_attrib[s].y = 0;
		sprite_attrib[s].sprite_pattern = s;
		sprite_attrib[s].color_code = 0xf;
		sprite_attrib[s].reserved = 0;
		sprite_attrib[s].early_clock = 1;
	}

	vdp_vwrite(sprite_attrib, SprAttrTable, TOTAL_SPRITES * sizeof(SPRITE_ATTRIBUTE));
	vdp_vwrite(sprite_set, SprPatTable, TOTAL_SPRITES * 8);

	movsprite(-1, -1, 0x0f);
}

void showsprite(bool show)
{
	static char last_y_value = 0;

	if (show && (sprite_attrib[0].y == SPRITE_TERMINATOR))
	{
		sprite_attrib[0].y = last_y_value;
	} else
	{
		if (sprite_attrib[0].y != SPRITE_TERMINATOR)
		{
			last_y_value = sprite_attrib[0].y = sprite_attrib[0].y;
			sprite_attrib[0].y = SPRITE_TERMINATOR;
		}
	}
}


/* 
 move off screen by using -1, -1
*/
void movsprite(int column, int row, int color)
{
	int x1, y1, sprite,x,y;
	char early_clock;


	if ((column == -1) || (row == -1))
	{
		x1 = 0;
		y1 = 0;
		early_clock = 1;
	} else
	{
		x1 = column * 16 + BOARD_START_X * 8;
		y1 = row    * 16 + BOARD_START_Y * 8 - 1;
		early_clock = 0;
	}
	char colors[] = {3, 2, 5, 4, 0xb, 0xa, 0xf};
	char temp[30];

	// 0 = Transparent
	// 1 = Black
	// 2 = Medium Green
	// 3 = Light Green
	// 4 = Dark Blue
	// 5 = Light Blue
	// 6 = Dark Red
	// 7 = Cyan
	// 8 = Medium Red
	// 9 = Light Red
	// A = Dark Yellow
	// B = Light Yellow
	// C = Dark Green
	// D = Magenta
	// E = Gray
	// F = White

	sprite = 0;
	for (x = x1; x < x1 + 32; x += 8)
	{
		for (y = y1; y < y1 + 32; y += 8)
		{
			sprite_attrib[sprite].x = x;
			sprite_attrib[sprite].y = y;
			sprite_attrib[sprite].color_code = color;
			sprite_attrib[sprite].early_clock = early_clock;

			sprite++;
			if (sprite == TOTAL_SPRITES)
				break;
		}
		if (sprite == TOTAL_SPRITES)
			break;

		y1 -= 15;
	}

	vdp_vwrite(sprite_attrib, SprAttrTable, TOTAL_SPRITES * sizeof(SPRITE_ATTRIBUTE));

}

unsigned char sprite_set[] =
{ // START SPRITES
	0b11111111,
	0b10000000,
	0b10000000,
	0b10000000,
	0b10000000,
	0b10000011,
	0b10000111,
	0b10000110,
	0b10000110,
	0b10000111,
	0b10000011,
	0b10000000,
	0b10000000,
	0b10000000,
	0b10000000,
	0b10000000,
	0b11111111,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000, // line 2347
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b11111111,
	0b00000001,
	0b00000001,
	0b00000001,
	0b00000001,
	0b11000001,
	0b11100001,
	0b01100001,
	0b01100001,
	0b11100001,
	0b11000001,
	0b00000001,
	0b00000001,
	0b00000001,
	0b00000001,
	0b00000001,
	0b11111111, // line 2348
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000,
	0b00000000  // line 2349
};
