package controllers

import (
	google_api "TaleHaven-Backend/modules/google-api"
	"github.com/gin-gonic/gin"
)

// BooksController
// Aici avem endpoint-urile noastre, adica niste rute/url-uri care au path diferit.
func BooksController(request *gin.Engine) {
	// EX: http://127.0.0.1:3000/search/books
	// adds prefix to routes
	loggedRoutes := request.Group("/search")
	loggedRoutes.GET("/books", searchBooks)    // GET /search/books
	loggedRoutes.GET("/book/:bookId", getBook) // GET /search/book/:id
	request.GET("/test", test)                 // GET /test
}

func searchBooks(context *gin.Context) {

	searchTerm := context.Query("query")

	if searchTerm == "" {
		context.JSON(400, gin.H{
			"message": "Empty query param",
		})

		return
	}

	booksList, err := google_api.SearchBooks(searchTerm)
	if err != nil {
		context.JSON(500, gin.H{
			"message": err.Error(),
		})

		return
	}

	context.JSON(200, gin.H{
		"booksList": booksList,
	})
}

func getBook(context *gin.Context) {
	bookId := context.Param("bookId")

	if bookId == "" {
		context.JSON(400, gin.H{
			"message": "Empty path param bookId",
		})

		return
	}

	booksList, _ := google_api.GetBooks(bookId)
	context.JSON(200, gin.H{
		"booksList": booksList,
	})

}

func test(context *gin.Context) {

	context.JSON(200, gin.H{
		"test": "Hello There",
	})
}
