import os
import sys
import pygame

# Local imports
my_modules_path = os.getcwd()+"/includes"
if sys.path[0] != my_modules_path:
    sys.path.insert(0, my_modules_path)
    

global done

def event_handler():
    
    done = False
    mousepos = (-1,-1)
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            print("Quit")
            done = True

        if event.type == pygame.MOUSEBUTTONUP:
          mousepos = pygame.mouse.get_pos()           

    return done, mousepos
