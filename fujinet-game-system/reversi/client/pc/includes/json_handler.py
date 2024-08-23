import json
import requests

class json_handler:
    
    def __init__(self,url):
        self.url = url
        self.json_data = None
        self.json_action_data = None
        self.data_change = False
        self.last_data = "@"
        self.table = ""
        self.board = ""
        self.connected = False
        self.my_name = ""
        
        self.key_last_result = 'l'
        self.key_players      = 'pl'
        self.key_valid_moves = 'vm'
        self.key_active_player = 'a'
        self.key_name = 'n'
        self.key_move = 'm'
        self.key_board = 'bd'
        self.key_timer = 'm'
        self.key_score = 'sc'
        self.key_color = 'c'
         
    def set_table(self, table):
        self.table = table
        
    def set_name(self, name):
        self.my_name = name
        
    def set_players(self, players):
        request = self.url+"/state?count="+str(players)+"&table="+self.table
        requests.get(request)
        
    def get_tables(self):
        try:
            request = self.url+"/tables"
            response = requests.get(request)
            self.table_data = json.loads(response.text)
            self.connected = True
        except Exception as e:
            print(f"error: request {request} -- {e}")
            self.connected = False
        return self.table_data 
        
    def refresh_data(self):
        try:
            request = self.url+"/state?player="+self.my_name+"&table="+self.table
            response = requests.get(request)
            self.json_data = json.loads(response.text)
            self.data_change = not (self.last_data == response.text)
            self.last_data = response.text;
            #if self.data_change:
            #    print(response.text)
            self.connected = True
        except Exception as e:
            print(f"error: request {request} -- {e}")
            self.connected = False
        return self.data_change 
    
    def get_board(self):
        if self.json_data == None or self.json_data[self.key_board] == None:
            return "."*64
        else:
            return self.json_data[self.key_board]
    
    def get_move_time(self):
        if self.json_data == None:
            return 0
        else:
            return self.json_data[self.key_timer]

    def put_move(self, row, col):
        success = True
        action = f":\"{row*8+col}\""
        try:
            request = self.url+"/move/"+action+"?player="+self.my_name+"&table="+self.table
            response = requests.get(request)
            self.json_action_data = json.loads(response.text)
            self.connected = True
        except:
            success = False
            self.connected = False
        return success
    
    def get_number_of_players(self):
        if self.json_data == None:
            return 0
        
        num = len(self.json_data[self.key_players])
        return num
    
    def get_name(self,player_num):
        try:
            name = self.json_data[self.key_players][player_num][self.key_name]
        except:
            name = "WAITING"
           
        return name
    
    
    def get_color(self, player_num):
        if self.json_data == None:
            return '.'

        if len(self.json_data[self.key_players]) > player_num-1:
            return '.'
        else:
            return self.json_data[self.key_players][player_num][self.key_color]
            
    def get_playing(self, player_num):
        if self.json_data == None:
            return -1
        
        player = self.json_data[self.key_active_player]
        return player
    
    def get_last_result(self):
        if self.json_data == None:
            return ""
        
        return self.json_data[self.key_last_result]
    
    def get_active_player(self):
        if self.json_data == None:
            return -1
        else:
            return self.json_data[self.key_active_player]
    
    def get_score(self, player_num):
        if self.json_data == None:
            score = 0
        else:
            try:
                score = self.json_data[self.key_players][player_num][self.key_score]
            except:
                score = 0
        return score
    
    def get_valid_moves(self):
        valid_moves = None
        
        moves = {}
        i = 0
        no_error = True
        while no_error:
            try:
                move = self.json_data[self.key_valid_moves][i][self.key_move]
                name = self.json_data[self.key_valid_moves][i][self.key_name]
                moves[move] = name
                i += 1
            except:
                no_error=False     
        
        return moves
    
    def is_valid_move(self, row, col):
        valid_move = False
        
        if self.json_data == None:
            return valid_move
        
        if self.json_data[self.key_valid_moves] != None:
            pos = str(row*8+col)
            for i in range(len(self.json_data[self.key_valid_moves])):
                if pos in self.json_data[self.key_valid_moves][i][self.key_move]:
                    valid_move = True
                    break
        
        return valid_move
            
    
