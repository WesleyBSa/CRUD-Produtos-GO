package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type Produto struct {
	ID         int     `json:"id,omitempty"`
	Nome       string  `json:"nome"`
	Preco      float64 `json:"preco"`
	Quantidade int64   `json:"quantidade"`
	Descricao  string  `json:"descricao"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "store.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS produto (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT,
		preco REAL,
		quantidade INT,
		descricao TEXT
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		log.Fatal(err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func adicionaOuAlteraProduto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var produto Produto

	err := r.ParseForm()
	if err != nil {
		log.Printf("Erro ao processar o formulário: %v", err)
		http.Error(w, "Erro ao processar o formulário", http.StatusInternalServerError)
		return
	}

	idStr := r.FormValue("id")
	if idStr != "" {
		produto.ID, err = strconv.Atoi(idStr)
		if err != nil {
			log.Println("ID inválido:", err)
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
	}

	produto.Nome = r.FormValue("nome")
	precoStr := r.FormValue("preco")
	quantidadeStr := r.FormValue("quantidade")
	produto.Preco, err = strconv.ParseFloat(precoStr, 64)
	if err != nil {
		log.Println("Erro ao converter preço:", err)
		http.Error(w, "Preço inválido", http.StatusBadRequest)
		return
	}

	produto.Quantidade, err = strconv.ParseInt(quantidadeStr, 10, 64)
	if err != nil {
		log.Println("Erro ao converter quantidade:", err)
		http.Error(w, "Quantidade inválida", http.StatusBadRequest)
		return
	}

	produto.Descricao = r.FormValue("descricao")

	log.Printf("Produto recebido: %+v", produto)

	if produto.Nome == "" {
		log.Println("Nome do produto é obrigatório")
		http.Error(w, "Nome do produto é obrigatório", http.StatusBadRequest)
		return
	}

	if produto.ID == 0 {
		_, err := db.Exec("INSERT INTO produto (nome, preco, quantidade, descricao) VALUES (?, ?, ?, ?)", produto.Nome, produto.Preco, produto.Quantidade, produto.Descricao)
		if err != nil {
			log.Printf("Erro ao adicionar produto: %v", err)
			http.Error(w, "Erro ao adicionar produto", http.StatusInternalServerError)
			return
		}
		log.Println("Produto adicionado com sucesso")
	} else {
		_, err := db.Exec("UPDATE produto SET nome = ?, preco = ?, quantidade = ?, descricao = ? WHERE id = ?", produto.Nome, produto.Preco, produto.Quantidade, produto.Descricao, produto.ID)
		if err != nil {
			log.Printf("Erro ao atualizar produto: %v", err)
			http.Error(w, "Erro ao atualizar produto", http.StatusInternalServerError)
			return
		}
		log.Println("Produto atualizado com sucesso")
	}

	http.Redirect(w, r, "/produtos-lista", http.StatusSeeOther)
}

func removeProduto(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM produto WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Erro ao remover produto", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/produtos-lista", http.StatusSeeOther)
}

func produtoFormulario(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("produto-formulario.html"))
	tmpl.Execute(w, nil)
}

func produtoAlteraFormulario(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)

	if r.Method == http.MethodPost {
		adicionaOuAlteraProduto(w, r)
		return
	}

	var produto Produto

	err = db.QueryRow("SELECT id, nome, preco, quantidade, descricao FROM produto WHERE id = ?", id).Scan(&produto.ID, &produto.Nome, &produto.Preco, &produto.Quantidade, &produto.Descricao)
	if err != nil {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("produto-altera-formulario.html"))
	tmpl.Execute(w, produto)
}

func produtosLista(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nome, preco, quantidade, descricao FROM produto")
	if err != nil {
		http.Error(w, "Erro ao listar produtos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var produtos []Produto

	for rows.Next() {
		var produto Produto
		if err := rows.Scan(&produto.ID, &produto.Nome, &produto.Preco, &produto.Quantidade, &produto.Descricao); err != nil {
			http.Error(w, "Erro ao ler produto", http.StatusInternalServerError)
			return
		}
		produtos = append(produtos, produto)
	}

	tmpl := template.Must(template.ParseFiles("produtos-lista.html"))
	tmpl.Execute(w, produtos)
}

func main() {
	initDB()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", home)
	http.HandleFunc("/adiciona-produto", adicionaOuAlteraProduto)
	http.HandleFunc("/remove-produto", removeProduto)
	http.HandleFunc("/produto-formulario", produtoFormulario)
	http.HandleFunc("/produto-altera-formulario", produtoAlteraFormulario)
	http.HandleFunc("/produtos-lista", produtosLista)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
