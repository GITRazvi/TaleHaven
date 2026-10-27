package anilist_api

import "strings"

// SearchMangaQuery returns a compact public metadata query with only the
// selected filter arguments. AniList treats a null search argument differently
// from an omitted argument, so optional fields must not be included as null.
func SearchMangaQuery(filters MangaSearchFilters) string {
	declarations := []string{"$page: Int!", "$perPage: Int!", "$sort: [MediaSort!]!"}
	arguments := []string{"type: MANGA", "sort: $sort"}
	if filters.Title != "" {
		declarations = append(declarations, "$search: String!")
		arguments = append(arguments, "search: $search")
	}
	if filters.Genre != "" {
		declarations = append(declarations, "$genre: String!")
		arguments = append(arguments, "genre: $genre")
	}
	if filters.Tag != "" {
		declarations = append(declarations, "$tag: String!")
		arguments = append(arguments, "tag: $tag")
	}
	if filters.Year != 0 {
		declarations = append(declarations, "$year: Int!")
		arguments = append(arguments, "seasonYear: $year")
	}
	if filters.Format != "" {
		declarations = append(declarations, "$format: MediaFormat!")
		arguments = append(arguments, "format: $format")
	}
	if filters.Status != "" {
		declarations = append(declarations, "$status: MediaStatus!")
		arguments = append(arguments, "status: $status")
	}

	return `query SearchManga(` + strings.Join(declarations, ", ") + `) {
  Page(page: $page, perPage: $perPage) {
    pageInfo {
      currentPage
      hasNextPage
      perPage
      total
    }
    media(` + strings.Join(arguments, ", ") + `) {
      ...MangaSearchFields
    }
  }
}

` + mangaSearchFieldsFragment
}

const mangaSearchFieldsFragment = `fragment MangaSearchFields on Media {
  id
  idMal
  title {
    romaji
    english
    native
    userPreferred
  }
  synonyms
  description(asHtml: false)
  coverImage {
    large
    medium
    color
  }
  format
  status
  chapters
  volumes
  genres
  averageScore
  popularity
  siteUrl
  isAdult
}`

// MangaDetailsQuery returns the public metadata used by a title-details page.
// It intentionally does not request any AniList user or list data.
const MangaDetailsQuery = `
query MangaDetails($id: Int!) {
  Media(id: $id, type: MANGA) {
    ...MangaSearchFields
    bannerImage
    startDate {
      year
      month
      day
    }
    endDate {
      year
      month
      day
    }
    countryOfOrigin
    source
    favourites
    rankings {
      rank
      type
      allTime
      context
    }
    tags {
      name
      rank
      isMediaSpoiler
    }
    staff {
      edges {
        role
        node {
          id
          name {
            full
            native
            userPreferred
          }
        }
      }
    }
    characters(page: 1, perPage: 10) {
      pageInfo {
        currentPage
        hasNextPage
        perPage
        total
      }
      edges {
        role
        node {
          id
          name {
            full
            native
            userPreferred
          }
          image {
            large
            medium
          }
        }
      }
    }
    externalLinks {
      id
      site
      url
      type
      language
    }
    relations {
      edges {
        relationType
        node {
          id
          type
          title {
            romaji
            english
            native
            userPreferred
          }
          coverImage {
            large
            medium
            color
          }
          format
          status
          siteUrl
        }
      }
    }
  }
}

fragment MangaSearchFields on Media {
  id
  idMal
  title {
    romaji
    english
    native
    userPreferred
  }
  synonyms
  description(asHtml: false)
  coverImage {
    large
    medium
    color
  }
  format
  status
  chapters
  volumes
  genres
  averageScore
  popularity
  siteUrl
  isAdult
}
`

// MangaCharactersQuery retrieves one lightweight page of story characters for
// a manga details page after its initial character set is displayed.
const MangaCharactersQuery = `
query MangaCharacters($id: Int!, $page: Int!, $perPage: Int!) {
  Media(id: $id, type: MANGA) {
    characters(page: $page, perPage: $perPage) {
      pageInfo {
        currentPage
        hasNextPage
        perPage
        total
      }
      edges {
        role
        node {
          id
          name {
            full
            native
            userPreferred
          }
          image {
            large
            medium
          }
        }
      }
    }
  }
}
`

// MangaDiscoveryQuery returns a paged public manga list for a home discovery
// section. The permitted ordering is selected by the backend, not by a browser
// supplied GraphQL query.
const MangaDiscoveryQuery = `
query MangaDiscovery($page: Int!, $perPage: Int!, $sort: [MediaSort!]!) {
  Page(page: $page, perPage: $perPage) {
    pageInfo {
      currentPage
      hasNextPage
      perPage
      total
    }
    media(type: MANGA, isAdult: false, sort: $sort) {
      ...MangaDiscoveryFields
    }
  }
}

fragment MangaDiscoveryFields on Media {
  id
  idMal
  title {
    romaji
    english
    native
    userPreferred
  }
  synonyms
  description(asHtml: false)
  coverImage {
    large
    medium
    color
  }
  format
  status
  chapters
  volumes
  genres
  averageScore
  popularity
  siteUrl
  isAdult
}
`

// MangaGenreQuery returns a public, popularity-ordered manga listing for one
// AniList genre. The genre is always supplied as a GraphQL variable.
const MangaGenreQuery = `
query MangaGenre($genre: String!, $page: Int!, $perPage: Int!) {
  Page(page: $page, perPage: $perPage) {
    pageInfo {
      currentPage
      hasNextPage
      perPage
      total
    }
    media(type: MANGA, isAdult: false, genre: $genre, sort: POPULARITY_DESC) {
      ...MangaGenreFields
    }
  }
}

fragment MangaGenreFields on Media {
  id
  idMal
  title {
    romaji
    english
    native
    userPreferred
  }
  synonyms
  description(asHtml: false)
  coverImage {
    large
    medium
    color
  }
  format
  status
  chapters
  volumes
  genres
  averageScore
  popularity
  siteUrl
  isAdult
}
`

const MangaByIDsQuery = `query MangaByIDs($ids: [Int!]!) { Page(perPage: 25) { media(id_in: $ids, type: MANGA) { ...MangaByIDsFields } } }
fragment MangaByIDsFields on Media { id title { romaji english native userPreferred } description(asHtml: false) coverImage { large medium color } format status chapters volumes genres averageScore popularity siteUrl isAdult }
`

const CharactersByIDsQuery = `query CharactersByIDs($ids: [Int!]!) { Page(perPage: 50) { characters(id_in: $ids) { id name { full native userPreferred } image { large medium } } } }`

const CharacterDetailsQuery = `
query CharacterDetails($id: Int!) {
  Character(id: $id) {
    id
    name { full native userPreferred }
    image { large medium }
    description(asHtml: false)
    gender
    dateOfBirth { year month day }
    age
    bloodType
    siteUrl
    favourites
    media(page: 1, perPage: 25, sort: POPULARITY_DESC, type: MANGA) {
      pageInfo { currentPage hasNextPage perPage total }
      edges {
        characterRole
        node {
          id
          type
          title { romaji english native userPreferred }
          coverImage { large medium color }
          format
          status
          siteUrl
        }
      }
    }
  }
}`
