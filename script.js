document.getElementById('produtoForm').addEventListener('submit', async function(event) {
    event.preventDefault();

    const nome = document.getElementById('nome').value;
    const preco = parseFloat(document.getElementById('preco').value);
    const quantidade = parseFloat(document.getElementById('quantidade').value);
    const descricao = document.getElementById('descricao').value;

    if (!nome || preco <= 0 || isNaN(preco)) {
        alert("Por favor, preencha corretamente os campos Nome e Preço.");
        return;
    }

    const produto = { nome, preco, quantidade, descricao };

    try {
        const response = await fetch('/adiciona-produto', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(produto)
        });

        if (!response.ok) {
            const errorMsg = await response.text();
            throw new Error('Erro ao cadastrar produto: ' + errorMsg);
        }

        window.location.href = '/produtos-lista';
    } catch (error) {
        alert(error.message);
    }
});

const productList = document.getElementById("product-list");
if (productList) {
    async function fetchProducts() {
        try {
            const response = await fetch("/produtos-lista");
            if (!response.ok) {
                throw new Error("Erro ao buscar produtos.");
            }

            const products = await response.json();

            productList.innerHTML = products.map((product) => `
                <div class="card mt-3">
                    <div class="card-body">
                        <h5 class="card-title">${product.nome}</h5>
                        <p class="card-text">${product.descricao}</p>
                        <p class="card-text">Preço: R$ ${product.preco.toFixed(2)}</p>
                        <p class="card-text">Quantidade: ${product.quantidade}</p>
                        <button class="btn btn-danger" onclick="deleteProduct(${product.id})">Deletar</button>
                    </div>
                </div>
            `).join("");
        } catch (error) {
            console.error("Erro ao buscar produtos:", error);
        }
    }

    fetchProducts();
}

async function deleteProduct(id) {
    try {
        const response = await fetch(`/remove-produto?id=${id}`, {
            method: "GET"
        });

        if (response.ok) {
            alert("Produto deletado com sucesso!");
            location.reload(); 
        } else {
            alert("Erro ao deletar produto.");
        }
    } catch (error) {
        console.error("Erro:", error);
        alert("Erro ao deletar produto.");
    }
}
document.getElementById('preco').addEventListener('input', function(e) {
    let value = e.target.value.replace(/[^0-9,]/g, ''); 
    value = value.replace(/,/g, '.'); 
    e.target.value = value;
});

