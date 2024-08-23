# Standard imports
import time
import sys
import os
os.environ['PYGAME_HIDE_SUPPORT_PROMPT'] = "hide"
import pygame

ip     = "fujinet-vm"
ip     = "FUJINET-VM.local"
#ip 	   = "localhost"


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

done = False

class Reversi:
    
    def __init__(self, ip):
        # Initialize pygame hooks
        pygame.init()
        pygame.display.set_caption('Reversi')
        pygame.font.init()
        
        self.clock = pygame.time.Clock()
      
        # built the url for the server website based on the ip/hostname supplied
        self.url = f'http://{ip}:8080'

        self.server = json_handler(self.url)
        
        # Set up the drawing window
        self.screen_width  = SCREEN_WIDTH
        self.screen_height = SCREEN_HEIGHT
        self.screen = pygame.display.set_mode([SCREEN_WIDTH, SCREEN_HEIGHT])
        
        pygame.display.set_caption(f'Fujinet Game Server Reversi: {self.url}')
        
        self._rows = 8
        self._cols = 8
        
        # this is the size of the board within the window
        self.game_offset_x = (self.screen_width  / (self._cols+2)) 
        self.game_offset_y = (self.screen_height / (self._rows+2))
        self.game_width    = (self.screen_width  / (self._cols+2))*self._cols
        self.game_height   = (self.screen_height / (self._rows+2))*self._rows
        
        # use these to find the row and column 
        self.row_multiplier = float(self.game_height) / self._rows
        self.col_multiplier = float(self.game_width)  / self._cols
        
        # the colours used
        self.game_background  = (0,  255,  0)
        self.text_color_black = (0,    0,  0)
        self.text_color_white = (255,255,255)
        self.inactive_player  = (128,128,128)
        
        # the fonts used
        self.font 		= pygame.font.SysFont('sans', 32, bold=True)
        self.font_mono  = pygame.font.SysFont('courier', 32)
        self.table = "no table"
        
        # clear the screen
        self.screen.fill(self.game_background)
        
        pygame.display.update()
        
    def set_player_color(self, player_num):
        player      = self.server.get_color(player_num)
        playerColor = self.inactive_player
        
        if player == 'B':
            playerColor = self.text_color_black
        if player == 'W':
            playerColor = self.text_color_white
        
        return playerColor
        
    
    def redraw_board(self):
        # just to always insure variables are initialized
        player1 = 0
        player2 = 1
        
        player1Color     = self.set_player_color(player1)
        player1backColor = self.game_background
        player2Color     = self.set_player_color(player2)
        player2backColor = self.game_background
        
        # find out who is playing -1=no one, 0=player 1, 1=player2
        active_player = self.server.get_active_player()

        # clear the screen
        self.screen.fill(self.game_background)
        
        # display the table we're on
        self.draw_string(f"{self.table}", 1, center=True, update=False, text_color=self.text_color_black,
                         text_background=self.game_background)
        
        #********** player 1 ************

        if active_player == player1:
            # inverse text for active player
            text       = player1backColor
            background = player1Color
        else:
            # normal text for the player that is waiting
            text       = player1Color
            background = player1backColor

        # print out name on the first line
        self.draw_string(self.server.get_name(player1), 0, text_color=text, text_background=background, update=False)
        
        #********** player 2 ************
        
        if active_player == player2:
            # inverse for active player
            text       = player2backColor
            background = player2Color
        else:
            # normal text for waiting player
            text       = player2Color
            background = player2backColor

        # print the name on the first line, but right justified
        self.draw_string(self.server.get_name(player2), 0, text_color=text, text_background=background,
                         update=False, right_justify=True)


        #********** message ************
        self.draw_string(self.server.get_last_result(), 2, text_color=self.text_color_black, text_background=self.game_background, update=False, center=True)


        #************** Print Timer
        if active_player >= 0:
            # someone is active
            if active_player == player1:
                # print the timer on the left side of screen
                self.draw_string(f"{self.server.get_move_time():3}", 2, update=False)
            else:
                # print the timer on the right side of the screen
                self.draw_string(f"{self.server.get_move_time():3}", 2, update=False, right_justify=True)
            

        # print the scores
        self.draw_string(f"{self.server.get_score(player1):3}", 10, update=False)           
        self.draw_string(f"{self.server.get_score(player2):3} ", 10, update=False, right_justify=True)           
            

        # Redraws the board 
        self.redraw_lines()
        self.redraw_cells()
        
        # if we're the current active player,
        # show all the valid moves
        if self.server.get_name(active_player) == self.myname:
            moves = self.server.get_valid_moves()
            self.draw_valid_moves(moves)
        
        # display everything we're drawn
        pygame.display.flip()

    # draws the row and column lines for the board
    def redraw_lines(self):
           
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
  
        
    # Redraws all the occupied cells in the board 
    def redraw_cells(self):
        
        for row in range(self._rows):
            for col in range(self._cols):
                if self.board[row*8+col] != '.':
                    self.draw_cell(row, col)
                    
    
    # draws a single piece at the row/col specified in the appropriate colour 
    def draw_cell(self, row, col):
        
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

    # show a grey circle where we clicked
    def draw_click(self, row, col):
        
        pygame.draw.circle(self.screen, (128, 128, 128),
                           (row * self.row_multiplier + (self.col_multiplier/2) + self.game_offset_x,
                            col * self.col_multiplier + (self.row_multiplier/2) + self.game_offset_y),
                            self.row_multiplier/2-2)
        pygame.display.update()
 
 
    # goes through the list of valid moves
    # and draws a hollow circle to indicate it is a valid move
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
        global done
        
        
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
            

        first_time = True
 
        while not done:
            
            # return the mouse position if clicked
            # done will equal false if the close button is clicked
            done, mouse_pos = event_handler()
            # was the mouse clicked?
            if mouse_pos != (-1,-1):
                # calculate the row and column and draw a grey circle
                row = (int) ((mouse_pos[0] - self.game_offset_x) / self.col_multiplier)
                col = (int) ((mouse_pos[1] - self.game_offset_y) / self.row_multiplier)
                self.draw_click(row,col)
                
                # if it was a valid move, then send it to the server.
                if self.server.is_valid_move(row,col):
                    self.server.put_move(row,col)
            
            # get the data from the server
            data_change = self.server.refresh_data()
            
            if not self.server.connected:
                print(f"{self.url} is down.")
                break
            
            # if the data hasn't changed since we last
            # called, don't do anything
            if (not data_change) and (not first_time):
                time.sleep(1)
                continue
  
            # time to do stuff
            
            self.board = self.server.get_board()
            
            # if the board has changed, then let the user know
            if self.board != last_board:
                self.beep()
                last_board = self.board
                        
            self.redraw_board()
            
            
            
            time.sleep(1)
            
            first_time = False
            
                
        print("Done.")
        pygame.quit()



    def get_string(self, line, centered=True):
        global done
        
        result = ""
        
        if done:
            getting_input = False
        else:
            getting_input = True
            
        while getting_input:
       
            # creating a loop to check events that 
            # are occurring
            for event in pygame.event.get():
                if event.type == pygame.QUIT:
                    print("Quit")
                    done = True
                    getting_input = False
                    
                # checking if keydown event happened or not
                if event.type == pygame.KEYDOWN:                    

                    if event.key == pygame.K_RETURN:
                        getting_input = False
                        break
                    
                    if event.key == pygame.K_BACKSPACE:
                        # erase the previous text
                        self.draw_string(result, line, center=centered, text_color=self.game_background)
                        # remove the last character that was input (if any)
                        result = result[:-1]
                    else:
                        result = result + event.unicode
                        
                    self.draw_string(result, line, center=centered)
                        
        return result
    
    
    
    def draw_string(self, message, line, x = 0, center=False, mono=False, update=True, text_color=None, text_background=None, right_justify=False):

        if text_color == None:
            text_color = self.text_color_black
            
        if text_background == None:
            text_background = self.game_background
                
        if mono:
            font = self.font_mono
        else:
            font = self.font
            
        text = font.render("X", True, text_color, text_background)
        textSize = text.get_rect()
        
        text = font.render(f"{message}", True, text_color, text_background)
        textRect = text.get_rect()

        
        if center:
            textRect.center = (self.screen_width/2, textSize.height*line)
        else:
            if right_justify:
                textRect = (self.screen_width-textRect.width, textSize.height*line)
            else:
                textRect = (textSize.width*x, textSize.height*line)

        self.screen.blit(text, textRect)
        if update:
            pygame.display.update()


    def get_name(self):
        self.screen.fill(self.game_background)
        
        line = 0
        self.draw_string(f"Your server is: {self.url}", line)
        
        line = 12
        self.draw_string("Enter your name: ", line, center=True, mono=True)
        
        line += 2
        return self.get_string(line)


    
    def get_table(self):
        self.screen.fill(self.game_background)
        tables = self.server.get_tables()
 
        line = 2
        self.draw_string(f"Table   Description", line, mono=True)
        
        line += 1
        
        for table in tables:
            
            self.draw_string(f"{table['t']:15}{table['n']}", line, mono=True)
            line += 1

        line += 4
        
        self.draw_string("What table do you want to sit at?", line, center=True, mono=True)
        line += 2  
        return self.get_string(line, centered=True)

if __name__ == '__main__':
    
    
    print(f"Your server is: {ip}")
    MY_GAME = Reversi(ip)
    myname  = MY_GAME.get_name()
    print(f"Your name is: {myname}") 
    mytable = MY_GAME.get_table()
    print(f"Your table is: {mytable}")
    MY_GAME.start(myname, mytable)
    