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
type statStruct struct {
	Name string `json:"name"`
}
type statsStruct struct {
	Base_Stat int `json:"base_stat"`
	Stat statStruct `json:"stat"`
}
type typeStruct struct {
	Type statStruct `json:"type"`
}
type requestPokemon struct {
	PokemonStats []statsStruct `json:"stats"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	Types []typeStruct `json:"Types"`
}
type Pokemon struct {
	name string
	basehp int
	height int
	weight int
	hp int
	attack int
	special_attack int
	special_defense int
	speed int
	defense int
	types []string

}

var commandMap map[string]cliCommand
var currentPage string
var previousPage string
var exploreLocation string
var catchPokemon string
var selectedPokemon string
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
		"inspect":{
			name:"inspect",
			description:"Inspect caught pokemon",
			callback: commandInspect,
		},
		"pokedex":{
			name:"pokedex",
			description:"List the pokedex",
			callback: commandPokedex,
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
		}else if cleanInput(strings.ToLower(input))[0] == "inspect"{
			selectedPokemon = cleanInput(strings.ToLower(input))[1]
			commandMap["inspect"].callback()
		}else if cleanInput(strings.ToLower(input))[0] == "pokedex"{
			commandMap["pokedex"].callback()
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
	const maxThrows = 3
	catchChance := 1.0 - float64(data.PokemonStats[0].Base_Stat)/255.0
	if catchChance < 0.10 {
		catchChance = 0.10
	}
	if catchChance > 0.90 {
		catchChance = 0.90
	}
	if rand.Float64() < catchChance {
		// fmt.Printf("data.PokemonStats %v\n",data.PokemonStats)
		// fmt.Printf("data.PokemonStats[0].Base_Stat %v\n",data.PokemonStats[0].Base_Stat)
		// fmt.Printf("data.PokemonStats[0].Stat.Name %v\n",data.PokemonStats[0].Stat.Name)
		var basehp int
		var attack int
		var defense int
		var speed int
		var special_attack int
		var special_defense int
		var types []string
		for i := 0; i <len(data.PokemonStats);i++{
			if data.PokemonStats[i].Stat.Name == "hp"{
				basehp = data.PokemonStats[i].Base_Stat
			}
			if data.PokemonStats[i].Stat.Name == "attack"{
				attack = data.PokemonStats[i].Base_Stat
			}
			if data.PokemonStats[i].Stat.Name == "speed"{
				speed = data.PokemonStats[i].Base_Stat
			}
			if data.PokemonStats[i].Stat.Name == "defense"{
				defense = data.PokemonStats[i].Base_Stat
			}
			if data.PokemonStats[i].Stat.Name == "special-attack"{
				special_attack = data.PokemonStats[i].Base_Stat
			}
			if data.PokemonStats[i].Stat.Name == "special-defense"{
				special_defense = data.PokemonStats[i].Base_Stat
			}
		}
		for i :=0;i<len(data.Types);i++{
			types = append(types,data.Types[i].Type.Name)
		}
		caughtPokemon[catchPokemon]=Pokemon{
			name: catchPokemon,
			basehp: basehp,
			speed: speed,
			attack:attack,
			defense:defense,
			special_attack:special_attack,
			special_defense:special_defense,
			height:data.Height,
			weight:data.Weight,
			types: types,
		}

		fmt.Printf("%s was caught!\n", catchPokemon)
		fmt.Printf("You may now inspect it with the inspect command.\n")
	}else {
		throw++
		catchChance += 0.05
		fmt.Printf("%s escaped!\n", catchPokemon)
	}
	return nil
}

func commandInspect() error {
	// fmt.Printf("inspect called%s",selectedPokemon)
	pokemon,err := caughtPokemon[selectedPokemon]
	if err {
		fmt.Printf("Name: %s\n",pokemon.name)
		fmt.Printf("Height: %d\n",pokemon.height)
		fmt.Printf("Weight: %d\n",pokemon.weight)
		fmt.Print("Stats:\n")
		fmt.Printf("  -hp: %d\n",pokemon.basehp)
		fmt.Printf("  -attack: %d\n",pokemon.attack)
		fmt.Printf("  -defense: %d\n",pokemon.defense)
		fmt.Printf("  -speical-attack: %d\n",pokemon.special_attack)
		fmt.Printf("  -special-defense: %d\n",pokemon.special_defense)
		fmt.Printf("  -speed: %d\n",pokemon.speed)
		fmt.Print("Types:\n")
		for i :=0;i<len(pokemon.types);i++{
			fmt.Printf("  - %s\n",pokemon.types[i])
		}
	}else {

		fmt.Printf("you have not caught that pokemon\n")
		
	}

	return nil

}

func commandPokedex() error {
	fmt.Printf("Your Pokedex:\n")
	for key := range caughtPokemon{
		fmt.Printf("  -%s\n",key)
	}
	return nil
}

