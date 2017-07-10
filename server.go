package main

import (
	"github.com/labstack/echo"
	"tomuss_server/src/controllers"
	"github.com/labstack/echo/middleware"
	"tomuss_server/src/metiers"
	"os"
	"tomuss_server/src/tools"
	"tomuss_server/src/models"
	"gopkg.in/olivere/elastic.v5"
	"time"
	"context"
	"encoding/json"
)

type Message struct {
	models.HeaderElasticsearch
	Source struct {
		Phrase string `json:"message" query:"message" form:"message"`
		Date string `json:"date" query:"date" form:"date"`
	} `json:"_source" form:"_source" query:"_source"`
}

func LoveMessage(c echo.Context) error {
	//Création du client
	client := tools.CreateElasticsearchClient()
	if client == nil {
		return c.JSON(403, models.JsonErrorResponse{Error: "Erreur lors de la connexion à notre base de donnée"})
	}

	results, err := client.Search().
	Index("lovemessage").
	Type("messages").
	Query(elastic.NewMatchQuery("date", time.Now().UTC().Format("2006-01-02"))).
	Pretty(true).
	Do(context.Background())

	if err != nil {
		return c.JSON(403, models.JsonErrorResponse{Error: "Erreur lors de la récupération du message du jour"})
	}

	//On récup le premier compte
	messageResult := results.Hits.Hits[0]
	message := new(models.Etudiant)

	bytes, err := json.Marshal(messageResult)

	//On parse le json en objet
	err_unmarshal := json.Unmarshal(bytes, &message)
	if err_unmarshal != nil {
		return c.JSON(403, models.JsonErrorResponse{Error: "Erreur lors de la récupération du message du jour"})
	}

	return c.JSON(200, message)
}

func main() {
	e := echo.New()

	//CORS
	e.Use(middleware.CORS())

	e.GET("/lovemessage", LoveMessage)

	//Association des routes
	//Définition des controllers
	etudiantController := new(controllers.EtudiantController)
	semestreController := new(controllers.SemestreController)

	//Gerant Controller sans JWT
	e.POST("/login", etudiantController.Login)
	e.POST("/register", etudiantController.Register)

	//Définition de l'api de base avec restriction JWT
	api := e.Group("/api")
	api.Use(middleware.JWT([]byte(metiers.GetSecretJwt())))

	//Api pour l'étudiant
	etudiantApi := api.Group("/etudiant")
	etudiantApi.PUT("/fcm", etudiantController.ChangeFcmToken)
	etudiantApi.PUT("/change/password", etudiantController.ChangePassword)
	etudiantApi.PUT("/change/informations", etudiantController.ChangeInformations)

	etudiantApi.GET("/semestres", semestreController.Find)
	etudiantApi.POST("/semestres", semestreController.Add)
	etudiantApi.PUT("/semestres/:id", semestreController.Update)
	etudiantApi.DELETE("/semestres/:id", semestreController.Remove)
	//semestreEtudiantApi.GET("", etudiantController.Profile)

	//Api pour les annonces
	/*annonceApi := api.Group("/annonce")
	annonceApi.GET("", annonceController.Find)
	annonceApi.GET("/:id", annonceController.Get)
	annonceApi.POST("", annonceController.Add)
	annonceApi.DELETE("/:id", annonceController.Delete)
	annonceApi.PUT("/:id", annonceController.Update)

	//Recherche des annonces par geolocation
	e.GET("/annonce/search/location", annonceController.SearchByLocation)*/

	go new(metiers.ScanRssMetier).Start()

	//On regarde comment on démarre le serveur
	env := os.Getenv("ENV")

	if env == "dev" {
		e.Logger.Fatal(e.Start(":1330"))
	} else{
		//e.AutoTLSManager.Cache = autocert.DirCache("/var/www/Golang-Projects/src/tomuss_server.cache")
		//e.Logger.Fatal(e.StartAutoTLS(":1330"))
		e.Logger.Fatal(e.Start(":1330"))
	}
}