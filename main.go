package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
	"io"
	"net/http"
	"encoding/json"
	"time"
	"github.com/DevankNassa/gopoke/internal/pokecache"
	"math/rand"
)
type cliCommand struct {
	name string
	description string
	callback func() error
}
type resultStruct struct{
	Name string `json:"name"`
	Url string `json:"url"`
}
type pokemonStruct struct{
	Pokemon resultStruct `json:pokemon`
}
type requestBody struct {
	Count int `json:"count"`
	Next string `json:"next"`
	Previous string `json:"previous"`
	Results []resultStruct `json:"results"`
}
type requestPokemonInArea struct {
	PokemonList []pokemonStruct `json:"pokemon_encounters"`
}
type statsStruct struct {
	Base_Stat int `json:"base_stat"`
}
type requestPokemon struct {
	PokemonStats []statsStruct `json:"stats"`
}
type Pokemon struct {
	name string
	basehp int

}

var commandMap map[string]cliCommand
var currentPage string
var previousPage string
var exploreLocation string
var catchPokemon string
var caughtPokemon = make(map[string]Pokemon)
var throw int

func init() {
	throw = 0
	currentPage = "https://pokeapi.co/api/v2/location-area"
	previousPage = ""
	commandMap = map[string]cliCommand{
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
							
		},
		"exit": {
			name: "exit",
			description: "Exit the Pokedex",
			callback: commandExit,
		},
		"map":{
			name: "map",
			description: "List the locaitons",
			callback: commandMapFunc,
		},
		"mapb":{
			name: "mapb",
			description: "List the locaitons in the previous page",
			callback: commandMapbFunc,
		},
		"explore":{
			name:"explore",
			description: "Explore pokemon in the area selected",
			callback: commandExplore,
		},
		"catch":{
			name:"catch",
			description:"Catch the pokemon",
			callback: commandCatch,
		},
	}
}


func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		if strings.ToLower(input) == "help"{
			// commandHelp()
			commandMap["help"].callback()
		}else if strings.ToLower(input) == "exit"{
			commandMap["exit"].callback()
		}else if strings.ToLower(input) == "map"{
			commandMap["map"].callback()
		}else if strings.ToLower(input) == "mapb"{
			commandMap["mapb"].callback()
		}else if cleanInput(strings.ToLower(input))[0] == "explore"{
			exploreLocation = cleanInput(strings.ToLower(input))[1]
			commandMap["explore"].callback()
		}else if cleanInput(strings.ToLower(input))[0] == "catch"{
			catchPokemon = cleanInput(strings.ToLower(input))[1]
			commandMap["catch"].callback()
		}else {
			fmt.Println("Unknown command")
		}
	}
}

func cleanInput(text string) []string {
	var initStr []string
	initStr = strings.Fields(text)
	return initStr
}

func commandExit() error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return fmt.Errorf("exit error")
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for command := range commandMap{
		fmt.Printf("%s: %s\n",commandMap[command].name,commandMap[command].description)
	}
	return nil
}

func commandMapFunc() error {
	callAPI(currentPage)
	return nil
}

func commandMapbFunc() error {
	if previousPage == ""{
		fmt.Println("you're on the first page")
	}else {
		callAPI(previousPage)
	}
	return nil
}

func callAPI(callURL string) error {
	cleanUpInterval := 10 * time.Second
	pokeCache := pokecache.NewCache(cleanUpInterval)
	item, found := pokeCache.Get(callURL)
	if !found{
		var processedData string
		res, err := http.Get(callURL)
		if err != nil { return err }
		body,err := io.ReadAll(res.Body)
		if res.StatusCode > 299 { return fmt.Errorf("response code over 299")}
		if err != nil { return err }
		defer res.Body.Close()
		var data requestBody
		errun := json.Unmarshal([]byte(body),&data)
		if errun != nil { return errun }
		for i := 0; i < len(data.Results);i++{
			processedData += fmt.Sprintf("%s\n",data.Results[i].Name)
		}
		pokeCache.Add(callURL,[]byte(processedData))
		fmt.Printf("%v",processedData)
		currentPage = data.Next
		previousPage = data.Previous
		return nil
	}
	fmt.Printf("%v",string(item))
	return nil
	
}

func commandExplore() error {
	res, err := http.Get("https://pokeapi.co/api/v2/location-area/"+exploreLocation)
	if err != nil { return err }
	body,err := io.ReadAll(res.Body)
	if res.StatusCode > 299 { return fmt.Errorf("response code over 299")}
	if err != nil { return err }
	defer res.Body.Close()
	var data requestPokemonInArea
	errun := json.Unmarshal([]byte(body),&data)
	if errun != nil { return errun }
	for i := 0; i < len(data.PokemonList);i++{
		fmt.Printf("%v\n",data.PokemonList[i].Pokemon.Name)
	}
	return nil
}

func commandCatch() error {
	fmt.Printf("Throwing a Pokeball at %v...\n",catchPokemon)
	res, err := http.Get("https://pokeapi.co/api/v2/pokemon/"+catchPokemon)
	if err != nil { return err }
	body,err := io.ReadAll(res.Body)
	if res.StatusCode > 299 { return fmt.Errorf("response code over 299")}
	if err != nil { return err }
	defer res.Body.Close()
	var data requestPokemon
	errun := json.Unmarshal([]byte(body),&data)
	if errun != nil { return errun }
	// fmt.Printf("base stat %v\n",data.PokemonStats[0].Base_Stat)	
	const maxThrows = 3
	catchChance := 1.0 - float64(data.PokemonStats[0].Base_Stat)/255.0
	if catchChance < 0.10 {
		catchChance = 0.10
	}
	if catchChance > 0.90 {
		catchChance = 0.90
	}
	if rand.Float64() < catchChance {
		caughtPokemon[catchPokemon]=Pokemon{
			name: catchPokemon,
			basehp:data.PokemonStats[0].Base_Stat,
		}
		fmt.Printf("%s was caught!\n", catchPokemon)
	}else {
		throw++
		fmt.Printf("%s escaped!\n", catchPokemon)
	}
	catchChance += 0.05
	return nil
}

