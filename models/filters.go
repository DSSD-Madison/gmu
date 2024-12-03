package models

type FilterOption struct {
	Label	string
	Count	int
}

type FilterCategory struct {
	Category	string
	Options		[]FilterOption
}

func Filters() []FilterCategory{
	return []FilterCategory{
		{
			Category: "Authors",
			Options: []FilterOption{
				{Label: "The United States Agency for International Development (USAID)", Count: 18},
				{Label: "Indikit", Count: 15},
				{Label: "Search for Common Ground (SFCG)", Count: 14},
				{Label: "Saferworld", Count: 5},
				{Label: "Mercy Corps / The United States Agency for International Development (USAID)", Count: 3},
				{Label: "Catholic Relief Services (CRS)", Count: 2},
				{Label: "International Rescue Committee (IRC)", Count: 2},
				{Label: "Mercy Corps", Count: 2},
				{Label: "Search for Common Ground (SFCG) / The United States Agency for International Development (USAID)", Count: 2},
				{Label: "The Asia Foundation", Count: 2},
			},
		},
		{
			Category: "File Type",
			Options: []FilterOption{
				{Label: "PDF", Count: 311},
				{Label: "MS WORD", Count: 51},
			},
		},
		{
			Category: "Source",
			Options: []FilterOption{
				{Label: "Alliance for Peacebuilding", Count: 97},
			},
		},
		{
			Category: "Keywords",
			Options: []FilterOption{
				{Label: "Development", Count: 21},
				{Label: "Capacity building", Count: 18},
				{Label: "Community", Count: 18},
				{Label: "Programs", Count: 16},
				{Label: "Relief", Count: 16},
				{Label: "Youth", Count: 8},
				{Label: "Access to justice", Count: 7},
				{Label: "Trainings", Count: 7},
				{Label: "Conflict resolution", Count: 6},
				{Label: "Media", Count: 6},
			},
		},
		{
			Category: "Region",
			Options: []FilterOption{
				{Label: "Global", Count: 31},
				{Label: "Democratic Republic of Congo (DRC)", Count: 7},
				{Label: "Burundi", Count: 6},
				{Label: "Uganda", Count: 5},
				{Label: "Afghanistan", Count: 4},
				{Label: "Morocco", Count: 4},
				{Label: "Bangladesh", Count: 3},
				{Label: "Guinea", Count: 3},
				{Label: "Kenya", Count: 3},
				{Label: "Nigeria", Count: 3},
			},
		},
	}
}
