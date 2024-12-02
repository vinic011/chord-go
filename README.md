# Chord DHT em Go

Este repositório contém uma implementação do protocolo Chord Distributed Hash Table, vulgo DHT, em Go para o exame da disciplina de CSC-27 de Processamento Distribuído. Nesse contexto, o protocolo Chord um dos precursores de DHT, que em função de sua simplicidade de conceito, é muito usado no cenário acadêmico de sistemas distribuídos. O Chord é um protocolo descentralizado, escalável e tolerante a falhas para armazenamento distribuído de pares chave-valor e busca eficiente.

## Pré-requisitos

Instalação da linguagem Go que pode ser feita de acordo com o tutorial na página oficial da mesma https://go.dev/doc/install.

## Instalação

1. Clone o repositório

```bash
git clone git@github.com:vinic011/chord-go.git
```

## Como usar

1. No diretório raiz, rode o comando

```bash
 go build -o myapp
```

2. Rode o programa compilado

```bash
./myapp
```

Vale notar que o programa está com um caso de teste para o caso de RingSize de tamanho 8 8, com um exemplo básico na main.go que demonstra a criação de uma rede de nós, inserção de novos nós, recuperação de valores e a impressão do estado da rede de forma a mostrar o correto funcionamento do algoritmo Chord simplificado com um exemplo concreto.

## Contribuições

Giuseppe Vicente Batista
Gustavo Pádua Beato
Nikollas da Silva Antes
Vinicius José de Menezes Pereira
