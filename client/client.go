package main

import (
	"context"
	"github.com/didikprabowo/go-grpc-json-mysql/blogpb"
	"github.com/didikprabowo/go-grpc-json-mysql/model"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"strconv"
)

// createArticle
func createArticle(c *gin.Context) {
	var article blogpb.Article

	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, cc := connectToServer()

	defer cc.Close()

	res, err := client.CreateArticle(context.Background(),
		&blogpb.CreateArticleRequest{Article: &article})

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"result": res,
		"code":   http.StatusCreated,
	})
}

// listArticle
func listArticle(c *gin.Context) {
	var page int64

	pageDefault := c.DefaultQuery("page", "1")
	page, err := strconv.ParseInt(pageDefault, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	client, cc := connectToServer()

	defer cc.Close()

	var articles []model.Article

	stream, err := client.ListArticle(context.Background(), &blogpb.ListArticleRequest{
		Page: page,
	})

	if err != nil {
		log.Fatal(err)
	}

	for {
		var article model.Article
		res, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		article.ID = res.GetArticle().GetId()
		article.Title = res.GetArticle().GetTitle()
		article.Body = res.GetArticle().GetBody()

		articles = append(articles, article)
	}
	code := http.StatusOK
	if len(articles) == 0 {
		code = http.StatusNoContent
	}
	c.JSON(code, gin.H{
		"result": articles,
		"code":   code,
	})
}

// connectToServerRPC
func connectToServer() (blogpb.ServiceNameClient, *grpc.ClientConn) {
	cc, err := grpc.Dial("0.0.0.0:5000", grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	client := blogpb.NewServiceNameClient(cc)
	return client, cc
}

// main
func main() {

	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("articles", listArticle)
	r.POST("article/create", createArticle)
	r.Run(":7000")
}
