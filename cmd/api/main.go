package main

import (
    "log"
    "net/http"
    "os"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"
    "context"
    "github.com/Akshay2904-bit/repomind/internal/ai"
    "github.com/Akshay2904-bit/repomind/internal/handler"
    "github.com/Akshay2904-bit/repomind/internal/repository"
    "github.com/Akshay2904-bit/repomind/internal/service"
    "github.com/redis/go-redis/v9"
)


func main() {
	//Load env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	//connect to postgresSQL
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Cannot connect to database: %v", err)
    }
    defer db.Close()

    rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("REDIS_URL"), // redis://localhost:6379
    })



	//initialize dependencies
    ollamaURL  := os.Getenv("OLLAMA_URL")   // e.g. http://localhost:11434
    embedModel := os.Getenv("EMBED_MODEL")  // e.g. nomic-embed-text
    chatModel  := os.Getenv("CHAT_MODEL")   // e.g. llama3.2
    embedder   := ai.NewEmbedder(ollamaURL, embedModel)
    llm        := ai.NewLLM(ollamaURL, chatModel)
    chunkRepo  := repository.NewChunkRepository(db)
    askSvc := service.NewAskService(rdb, embedder, chunkRepo, llm)
    h := handler.NewHandler(embedder, llm, chunkRepo, askSvc)


        // Set up HTTP router
    r := chi.NewRouter()
    r.Use(middleware.Logger)    // logs every request
    r.Use(middleware.Recoverer) // recovers from panics
 
    r.Post("/repos/{repo}/index", h.IndexRepo)  // trigger scan + embed
    r.Post("/ask", h.AskQuestion)               // RAG query
    r.Get("/repos/{repo}/files", h.ListFiles)   // list indexed files
 
    log.Printf("Server starting on port %s", os.Getenv("PORT"))
    log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), r))

    
}
