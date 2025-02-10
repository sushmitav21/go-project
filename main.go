package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const searchAPI = "https://api.apilayer.com/spoonacular/food/menuItems/search?query="
const recipeAPI = "https://api.apilayer.com/spoonacular/recipes/%d/information"

func loadAPIKey() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API key not found in environment variables")
	}
	return apiKey
}

type SearchResponse struct {
	MenuItems []MenuItem `json:"menuItems"`
}

type MenuItem struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	RestaurantChain string    `json:"restaurantChain"`
	Image           string    `json:"image"`
	Nutrition       Nutrition `json:"nutrition"`
}
type Nutrition struct {
	//	Nutrients []Nutrient `json:"nutrients"`
	Calories float32 `json:"calories"`
	Fat      string  `json:"fat"`
	Protein  string  `json:"protein"`
	Carbs    string  `json:"carbs"`
}

type RecipeResponse struct {
	Instructions string `json:"instructions"`
	GlutenFree   bool   `json:"glutenFree"`
	ReadyInMins  int    `json:"readyInMinutes"`
	Servings     int    `json:"servings"`
	vegan        bool   `json:"vegan"`
}

// Fetching menu items from API 1
func searchMenuItems(query string) ([]MenuItem, error) {
	apiKey := loadAPIKey()
	//	fmt.Println("✅ API Key Loaded Successfully")
	url := fmt.Sprintf("%s%s&addMenuItemInformation=true", searchAPI, query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("apikey", apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized: check if the API key is correct and properly included in the header")
	}

	var searchResponse SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}
	//fmt.Println(searchResponse.MenuItems)
	return searchResponse.MenuItems, nil
}

// API 2
func fetchRecipe(id int) (RecipeResponse, error) {
	apiKey := loadAPIKey()
	//fmt.Println("✅ API Key Loaded Successfully")
	url := fmt.Sprintf(recipeAPI, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RecipeResponse{}, fmt.Errorf("Error creating request: %w", err)
	}
	req.Header.Add("apikey", apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	var recipeResponse RecipeResponse
	err = json.NewDecoder(resp.Body).Decode(&recipeResponse)
	return recipeResponse, nil
}

// CLI command function
func runCLI(cmd *cobra.Command, args []string) {
	query, _ := cmd.Flags().GetString("query")

	items, err := searchMenuItems(query)
	if err != nil {
		log.Fatal("Error fetching menu items:", err)
	}
	fmt.Println("\n🍔✨ Browse for more information on the Menu Items Found in the restaurants near you..:✨🍔")
	fmt.Println("--------------------------------------------------")
	for _, item := range items {
		// Print each item in a visually attractive format
		fmt.Printf("\033[1;36m%3d - %s\033[0m\n", item.ID, item.Title)
		fmt.Printf("\033[1;32m   Restaurant: \033[0m%s\n", item.RestaurantChain)
		//	fmt.Printf("\033[1;33m   Calories: \033[0m%.0f kcal\n", item.Nutrition.Calories)
		fmt.Println("--------------------------------------------------")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nPlease enter the ID of the item you want to know more about: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	selectedID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("❌ Invalid ID. Please enter a valid number.")
		return
	}

	var selectedItem MenuItem
	for _, item := range items {
		if item.ID == selectedID {
			selectedItem = item
			break
		}
	}

	if selectedItem.ID == 0 {
		fmt.Println("❌ No matching item found with the specified ID.")
		return
	}

	recipeDetails, err := fetchRecipe(selectedItem.ID)
	if err != nil {
		log.Fatal("Error fetching recipe details:", err)
	}

	fmt.Printf("\n🍔 Found: %s at %s\n", selectedItem.Title, selectedItem.RestaurantChain)
	fmt.Println("🖼️ Image:", selectedItem.Image)
	fmt.Printf("🔥 Calories in %s: %.0f kcal\n", selectedItem.Title, selectedItem.Nutrition.Calories)
	fmt.Printf("🔥 Carbs in %s: %s \n", selectedItem.Title, selectedItem.Nutrition.Carbs)
	fmt.Printf("🔥 Protein in %s: %s \n", selectedItem.Title, selectedItem.Nutrition.Protein)
	fmt.Printf("🔥 Fats in %s: %s \n", selectedItem.Title, selectedItem.Nutrition.Fat)

	fmt.Println("\n🌱 Is it Vegan? ", getVeganStatus(recipeDetails))
	fmt.Println("🍞 Is it Gluten-Free? ", getGlutenFreeStatus(recipeDetails))
	fmt.Printf("⏳ Ready in: %d minutes\n", recipeDetails.ReadyInMins)
	fmt.Printf("🍽️ Servings: %d\n", recipeDetails.Servings)

	fmt.Print("\nDo you want the recipe? (yes/no): ")
	recipeResponse, _ := reader.ReadString('\n')
	recipeResponse = strings.TrimSpace(recipeResponse)

	if strings.EqualFold(recipeResponse, "yes") {
		fmt.Println("\n🥗 Recipe Instructions:")
		fmt.Println(recipeDetails.Instructions)
	} else {
		fmt.Println("No recipe requested.")
	}
}

func getVeganStatus(recipe RecipeResponse) string {
	if recipe.vegan {
		return "Yes ✅"
	}
	return "No ❌"
}

func getGlutenFreeStatus(recipe RecipeResponse) string {
	if recipe.GlutenFree {
		return "Yes ✅"
	}
	return "No ❌"
}

func main() {

	rootCmd := &cobra.Command{
		Use:   "foodsearch",
		Short: "Search fast food menu items",
		Run:   runCLI,
	}

	rootCmd.Flags().StringP("query", "q", "", "Food item to search (required)")

	_ = rootCmd.MarkFlagRequired("query")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
