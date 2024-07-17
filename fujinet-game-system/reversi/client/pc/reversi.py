# Standard imports
import time
import sys
import os
os.environ['PYGAME_HIDE_SUPPORT_PROMPT'] = "hide"
import pygame


# Local imports
my_modules_path = os.getcwd()+"/includes"
if sys.path[0] != my_modules_path:
    sys.path.insert(0, my_modules_path)
    
from json_handler import *
from event_handler import *

global done


class Reversi(object):
    
    # Initialize pygame hooks
    pygame.init()
    pygame.display.set_caption('Reversi')
    pygame.font.init()
    
    clock = pygame.time.Clock()

    url = 'http://localhost:8080'

    server = json_handler(url)

    
    server.set_name("TechCowboy")
    server.set_table("bot2a")
    #server.set_players(2)
    
    server.refresh_data()

    if not server.connected:
        print(f"{url} is down")
        exit(-1)
        
    # **********************************************
    # ALL STATIC VARIABLES HAVE BEEN CALCULATED
    # **********************************************

    first_time = True
    play_again = True
    done = False
    while play_again and not done:
        
        update_screen = 1
      
        while not done:
            event_handler()
            data_change = server.refresh_data()
            
            if not server.connected:
                print(f"{url} is down.")
                break
            
            if (not data_change) and (not first_time):
                time.sleep(0.5)
                continue
            
            board = server.get_board()
            
            pos = 0
            for y in range(8):
                for x in range(8):
                    print(board[y*8 + x], end='')
                print()
            
            
            time.sleep(1)
            
            first_time = False
        
            
    #end while play again
    print("Done.")
    pygame.quit()

if __name__ == '__main__':
    MY_GAME = Reversi()
    