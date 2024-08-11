import json
import requests



class json_handler:
    
    def __init__(self,url):
        self.url = url
        self.json_data = []
        self.data_change = False
        self.last_data = "@"
        self.my_name=""
        self.table = ""
        self.players = "pl"
        self.name = "n"
        self.num_players = "p"
        self.hand = "h"
        self.purse = "p"
        self.bet = 'b'
        self.activePlayer = 'a'
        self._round = 'r'
        self.pot = 'p'
        self.lastResult = 'l'
        self.validMoves = 'v'
        self.tables=[]
        
        return
    
    def get_tables(self):
        request=self.url+"/tables"
        print(f"get_tables: {request}")
        response = requests.get(request)
        self.tables = json.loads(response.text)
        print(self.tables)
        
    
    def set_table(self, table):
        self.table = table
        print(f"Set table {table}")
        
    def set_name(self, name):
        self.my_name = name
        
    def set_players(self, players):
        request=self.url+"/state?count="+str(players)+"&table="+self.table+"&player="+self.my_name
        print(f"set_players: {request}")
        requests.get(request)
        
    def refresh_data(self):
        request = self.url+"/state?table="+self.table+"&player="+self.my_name
        print(f"refresh: {request}")
        try:
            response = requests.get(request)

            self.json_data = json.loads(response.text)
            self.data_change = not (self.last_data == self.json_data)
            self.last_data = self.json_data;
            if self.data_change:
                print(response.text)
            self.connected = True
        except:
            self.connected = False
        return self.data_change 
    
    def send_action(self, action):
        success = True
        try:
            response = requests.get(self.url+"/move/"+action)
            print(f"***send_action: {action}")
            self.json_data = json.loads(response.text)
            
            self.connected = True
        except:
            success = False
            self.connected = False
        return success
    
    def get_number_of_players(self):
        num = len(self.json_data[self.players])
        return num
    
    def get_name(self,player_num):
        return self.json_data[self.players][player_num][self.name]
        
    def get_hand(self,player_num):
        return self.json_data[self.players][player_num][self.hand]
    
    def get_purse(self,player_num):
        return self.json_data[self.players][player_num][self.purse]
    
    def get_bet(self,player_num):
        return self.json_data[self.players][player_num][self.bet]
    
    def get_playing(self, player_num):
        player = self.json_data[self.activePlayer]
        return player
    
    def get_fold(self, player_num):
        return self.json_data[self.players][player_num][self.hand] == "??"
    
    def get_round(self):
        return self.json_data[self._round]
    
    def get_pot(self):
        return self.json_data[self.pot]
    
    def get_last_result(self):
        return self.json_data[self.lastResult]
    
    def get_active_player(self):
        return self.json_data[self.activePlayer]
    
    def get_valid_buttons(self):
        valid_moves = None
        
        moves = {}
        i = 0
        no_error = True
        while no_error:
            try:
                move = self.json_data[validMoves][i][self.move]
                name = self.json_data[validMoves][i][self.name]
                moves[self.move] = name
                i += 1
            except:
                no_error=False     
        
        return moves
    
    
