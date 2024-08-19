# Standard imports
import time
import sys
import os
os.environ['PYGAME_HIDE_SUPPORT_PROMPT'] = "hide"
import pygame

ip     = "fujinet-vm"
ip     = "FUJINET-VM.local"
myname = "VM-TC"
mytable= "green"

ip = "localhost"


# Local imports
my_modules_path = os.getcwd()+"/includes"
if sys.path[0] != my_modules_path:
    sys.path.insert(0, my_modules_path)
    
from json_handler import *
from event_handler import *

SCREEN_HEIGHT    = 800
SCREEN_WIDTH     = 800
BACKGROUND_COLOR = '#696969'
GAME_COLOR       = '#006000'




class Reversi:
    
    def __init__(self, ip):
        # Initialize pygame hooks
        pygame.init()
        pygame.display.set_caption('Reversi')
        pygame.font.init()
        
        self.clock = pygame.time.Clock()
      
        # built the url for the server website based on the ip supplied
        self.url = f'http://{ip}:8080'

        self.server = json_handler(self.url)
        
        # Set up the drawing window
        self.screen_width  = SCREEN_WIDTH
        self.screen_height = SCREEN_HEIGHT
        self.screen = pygame.display.set_mode([SCREEN_WIDTH, SCREEN_HEIGHT])
        
        pygame.display.set_caption(f'Fujinet Game Server Reversi: {self.url}')
        
        self._rows = 8
        self._cols = 8
        
        self.game_offset_x = (self.screen_width  / (self._cols+2)) 
        self.game_offset_y = (self.screen_height / (self._rows+2))
        self.game_width    = (self.screen_width  / (self._cols+2))*self._cols
        self.game_height   = (self.screen_height / (self._rows+2))*self._rows
        
        self.row_multiplier = float(self.game_height) / self._rows
        self.col_multiplier = float(self.game_width)  / self._cols
        
        self.game_background  = (0,  255,  0)
        self.text_color_black = (0,    0,  0)
        self.text_color_white = (255,255,255)
        
        self.font = pygame.font.Font('freesansbold.ttf', 32)
        self.table = "no table"
        
        self.screen.fill(self.game_background)
        pygame.display.update()
        
    
    def redraw_board(self):
        player1Color = self.text_color_black
        player1backColor = self.game_background
        player2Color = self.text_color_white
        player2backColor = self.game_background
        
        self.screen.fill(self.game_background)
        
        # table
        text = self.font.render(f"{self.table}", True, self.text_color_white, self.game_background)
         
        # create a rectangular object for the
        # text surface object
        textRect = text.get_rect()
         
        # set the center of the rectangular object.

        text_rect = text.get_rect(center=(self.screen_width/2, textRect.height))
        self.screen.blit(text, text_rect)

        
        #********** player 1 name ************

        if self.server.get_color(0) == 'B':
            player1Color = self.text_color_black
            player1backColor = self.game_background

        if self.server.get_color(0) == 'W':
            player1Color = self.text_color_white
            player1backColor = self.game_background
 
        if self.ap == 0:
            text = self.font.render(self.server.get_name(0), True, player1backColor, player1Color)
        else:
            text = self.font.render(self.server.get_name(0), True, player1Color, player1backColor)
         
        # create a rectangular object for the
        # text surface object
        textRect = text.get_rect()
         
        # set the center of the rectangular object.
        textRect.center = (textRect.width, textRect.height)
        self.screen.blit(text, textRect)

        #********** player 2 name ************
        
        if self.server.get_color(1) == 'B':
            player2Color = self.text_color_black
            player2backColor = self.game_background

        if self.server.get_color(1) == 'W':
            player2Color = self.text_color_white
            player2backColor = self.game_background

        
        if self.ap == 1:
            text = self.font.render(self.server.get_name(1), True, player2backColor, player2Color)
        else:
            text = self.font.render(self.server.get_name(1), True, player2Color, player2backColor)
            

         
        # create a rectangular object for the
        # text surface object
        textRect = text.get_rect()
         
        # set the center of the rectangular object.
        textRect.center = (self.screen_width - textRect.width, textRect.height)
        self.screen.blit(text, textRect)

        #********** message ************

        text = self.font.render(self.server.get_last_result(), True, self.text_color_white, self.game_background)
         
        # create a rectangular object for the
        # text surface object
        textRect = text.get_rect()
         
        # set the center of the rectangular object.

        text_rect = text.get_rect(center=(self.screen_width/2, textRect.height*2))
        self.screen.blit(text, text_rect)


        #************** Print Timer

        if self.ap >= 0:
            text = self.font.render(f"{self.server.get_move_time()}", True, self.text_color_black, self.game_background)
            score1 = self.font.render(f"{self.server.get_score(0)}", True, player1Color, self.game_background)
            score2 = self.font.render(f"{self.server.get_score(1)}", True, player2Color, self.game_background)
            textRect = text.get_rect()
            scoreRect1 = score1.get_rect()
            scoreRect2 = score2.get_rect()
            if self.ap == 0:
                textRect.center = (textRect.width, textRect.height*2)
            else:
                textRect.center = (self.screen_width - textRect.width, textRect.height*2)
            
            scoreRect1.center = (textRect.width, textRect.height*10)
            scoreRect2.center = (self.screen_width - textRect.width, textRect.height*10)
            self.screen.blit(text, textRect)
            self.screen.blit(score1, scoreRect1)
            self.screen.blit(score2, scoreRect2) 

        ''' Redraws the board '''
        self.redraw_lines()
        self.redraw_cells()
        
        if self.server.get_name(self.ap) == self.myname:
            moves = self.server.get_valid_moves()
            self.draw_valid_moves(moves)
                    
        pygame.display.update()

    def redraw_lines(self):
        ''' Redraws the board's lines '''
           
        # Draw the horizontal lines first
        for row in range(0, self._rows+1):
            pygame.draw.line(self.screen, pygame.Color("black"),
                             (row * self.row_multiplier+self.game_offset_x, self.game_offset_y),
                             (row*self.row_multiplier+self.game_offset_x,   self.game_height+self.game_offset_y), 1)

        # Draw the column lines next
        for col in range(0, self._cols+1):
            pygame.draw.line(self.screen, pygame.Color("black"),
                             (self.game_offset_x,                 col * self.col_multiplier+self.game_offset_y),
                             (self.game_width+self.game_offset_x, col * self.col_multiplier+self.game_offset_y), 1)
  
        

    def redraw_cells(self):
        ''' Redraws all the occupied cells in the board '''
        for row in range(self._rows):
            for col in range(self._cols):
                if self.board[row*8+col] != '.':
                    self.draw_cell(row, col)
                    
                
    def draw_cell(self, row, col):
        
        ''' Draws the specified cell '''
        if self.board[row * 8 + col] == 'B':
            pygame.draw.circle(self.screen, (0, 0, 0),
                               (row * self.row_multiplier + (self.col_multiplier/2) + self.game_offset_x,
                                col * self.col_multiplier + (self.row_multiplier/2) + self.game_offset_y),
                                self.row_multiplier/2-2)
        else:
            pygame.draw.circle(self.screen, (255, 255, 255),
                               (row * self.row_multiplier + (self.col_multiplier/2) + self.game_offset_x,
                                col * self.col_multiplier + (self.col_multiplier/2) + self.game_offset_y),
                                self.col_multiplier/2-2)

    def draw_click(self, row, col):
        

        pygame.draw.circle(self.screen, (128, 128, 128),
                           (row * self.row_multiplier + (self.col_multiplier/2) + self.game_offset_x,
                            col * self.col_multiplier + (self.row_multiplier/2) + self.game_offset_y),
                            self.row_multiplier/2-2)
        pygame.display.update()
 
 

    def draw_valid_moves(self, moves):
        for key in moves.keys():
            try:
                pos = int(key)
            except:
                continue
            row = pos // 8
            col = pos - row * 8
            pygame.draw.circle(self.screen, (128, 128, 128),
                               (row * self.row_multiplier + (self.col_multiplier/2) + self.game_offset_x,
                                col * self.col_multiplier + (self.col_multiplier/2) + self.game_offset_y),
                                self.col_multiplier/2-2, 2)
        pygame.display.update()
    
    def beep(self):
        print("\a")

    def start(self, myname, table):
        done = False
        
        self.beep()
        self.myname = myname
        self.table = table
        self.server.set_name(self.myname)
        self.server.set_table(self.table)
        
        # Initial Game Settings
        self._rows = 8
        self._columns = 8

        self.server.refresh_data()
        last_board=""

        if not self.server.connected:
            print(f"{self.url} is down")
            exit(-1)
            
        # **********************************************
        # ALL STATIC VARIABLES HAVE BEEN CALCULATED
        # **********************************************

        first_time = True
 
        while not done:
            done, mouse_pos = event_handler()
            data_change = self.server.refresh_data()
            
            if not self.server.connected:
                print(f"{self.url} is down.")
                break
            
            if (not data_change) and (not first_time):
                time.sleep(1)
                continue
            
            self.board = self.server.get_board()
            if self.board != last_board:
                self.beep()
                last_board = self.board
            
            timer = self.server.get_move_time()
            self.ap = self.server.get_active_player()
            
            self.redraw_board()
            
            if mouse_pos != (-1,-1):
                row = (int) ((mouse_pos[0] - self.game_offset_x) / self.col_multiplier)
                col = (int) ((mouse_pos[1] - self.game_offset_y) / self.row_multiplier)
                self.draw_click(row,col)
                if self.server.is_valid_move(row,col):
                    self.server.put_move(row,col)
            
            time.sleep(1)
            
            first_time = False
            
                
        print("Done.")
        pygame.quit()
        
    def get_string(self):
        result = ""
        getting_input = True
        while getting_input:
       
            # creating a loop to check events that 
            # are occurring
            for event in pygame.event.get():
                    
                # checking if keydown event happened or not
                if event.type == pygame.KEYDOWN:                    

                    if event.key == pygame.K_RETURN:
                        getting_input = False
                        break
                    
                    if event.key == pygame.K_BACKSPACE:
                        result = result[:-1]
                    else:
                        result = result + event.unicode
                        
        print(f"result = '{result}'")
        return result
    
        
    def get_name(self):
        self.screen.fill(self.game_background)
        print("\nEnter your name: ", end='')
        return self.get_string()

    
    def get_table(self):
        self.screen.fill(self.game_background)
        tables = self.server.get_tables()
        for table in tables:
            print(f"{table['t']}\t\t{table['n']}")
        
        
        print("\nEnter your table: ", end='')
        return self.get_string()

if __name__ == '__main__':
    
    
    print(ip)
    MY_GAME = Reversi(ip)
    myname  = MY_GAME.get_name()
    print(f"Your name is: {myname}") 
    mytable = MY_GAME.get_table()
    print(f"Your table is: {mytable}")
    MY_GAME.start(myname, mytable)
    