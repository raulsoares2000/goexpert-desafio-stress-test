Para executar o programa, dê o seguinte comando:
go run <nome da sua build> charge -u=<url> -r=<número de requests> -c=<número de chamadas simultâneas>

Ou, no docker:
docker run <sua imagem docker> charge -u=<url> -r=<número de requests> -c=<número de chamadas simultâneas>

Equivalência das flags:
-u = --url
-r = --requests
-c = --concurrency
