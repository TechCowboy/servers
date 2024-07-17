# Fujinet Game System Reversi

![Reversi](../images/reversi.png)

Special thanks for Thomas Cherryhomes for providing the 
initial user interace and introducing me to Magellan which 
I used to create the current board screen and sprites.

If you're interested in vintage/retro computing, on various
platforms, you *need* to check out the Fujinet at 
https://fujinet.online 


## Set up requirements

You need z88dk compiler 
https://github.com/z88dk/z88dk

## Make options

make clean - remove previous compiled objects<br/>
make       - build either the NABU homebrew or NABU CP/M executiable<br/>

## Graphics interface

The game board was created using Magellian; a TI-99/4a tile editor.
https://github.com/Rasmus-M/magellan

I wrote mag2c.py in Python 3 with takes the output of Magellian and replaces the
C source code for board.c, charset.c and spriteset.c -- a huge timesaver!





