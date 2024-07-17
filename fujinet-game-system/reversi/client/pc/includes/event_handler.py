import os
import sys
import pygame

# Local imports
my_modules_path = os.getcwd()+"/includes"
if sys.path[0] != my_modules_path:
    sys.path.insert(0, my_modules_path)
    

global done

def event_handler():
    
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            done = True

            

