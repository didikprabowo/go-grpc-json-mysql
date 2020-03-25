package main

import (
	"context"
	"fmt"
	"github.com/didikprabowo/go-grpc-json-mysql/blogpb"
	"github.com/didikprabowo/go-grpc-json-mysql/model"
	"github.com/didikprabowo/go-grpc-json-mysql/server/article/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
)

type server struct{}

var repo repository.ArticleRepository

func init() {
	repo = repository.NewMySQL()
}

// CreateArticle
func (*server) CreateArticle(ctx context.Context, req *blogpb.CreateArticleRequest) (*blogpb.CreateArticleResponse, error) {

	article := req.GetArticle()

	newArticle := model.Article{
		Title: article.GetTitle(),
		Body:  article.GetBody(),
	}

	res, err := repo.Create(ctx, newArticle)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("internal error : %v", err))
	}

	return &blogpb.CreateArticleResponse{
		Article: &blogpb.Article{
			Id:    res,
			Title: article.GetTitle(),
			Body:  article.GetBody(),
		},
	}, nil
}

// ListArticle
func (*server) ListArticle(req *blogpb.ListArticleRequest, stream blogpb.ServiceName_ListArticleServer) error {

	var perPage int64 = 10
	var start, end int64

	if req.GetPage() < 1 {
		start = 0
	} else {
		start = (req.GetPage() - 1) * perPage
	}

	end = perPage

	la, err := repo.List(context.Background(), start, end)

	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("internal error : %v", err))
	}

	for _, key := range la {
		stream.Send(&blogpb.ListArticleResponse{Article: &blogpb.Article{
			Id:    key.ID,
			Title: key.Title,
			Body:  key.Body,
		}})
	}

	return nil
}

// main
func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	list, err := net.Listen("tcp", "0.0.0.0:5000")

	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	blogpb.RegisterServiceNameServer(s, &server{})

	go func() {
		fmt.Println("Starting server..")
		if err := s.Serve(list); err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("stop the server..")
	s.Stop()
	fmt.Println("stop the listener..")
	list.Close()
}
