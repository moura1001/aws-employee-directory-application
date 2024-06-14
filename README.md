# AWS Employee directory application

Implementação em Go do projeto exemplo do curso AWS Technical Essentials. O curso aborda os principais aspectos de computação em
nuvem e cobre os conceitos fundamentais da AWS relacionados a computação, banco de dados, armazenamento, rede, monitoramento e
segurança. O projeto desenvolvido trata de uma aplicação web para gerenciamento de funcionários fictícios de uma empresa com deploy no
EC2 numa implantação multi-AZ visando alta disponibilidade, DynamoDB para persistência dos dados de cadastro e S3 para armazenamento das fotos de perfil dos funcionários.
Além disso, também é exercitado o monitoramento e escalabilidade da aplicação utilizando CloudWatch, EC2 Auto Scaling e Elastic Load Balancing.

![0-diagrama-arquitetura.png](docs/images/0-diagrama-arquitetura.png?raw=true "Diagrama de arquitetura AWS")

# Screenshots

Estado inicial da aplicação, sem nenhum dado:

Tela Inicial:

![1-home-empty.png](docs/images/1-home-empty.png?raw=true "Tela inicial da aplicação")

Bucket S3:

![2-bucket-s3-empty.png](docs/images/2-bucket-s3-empty.png?raw=true "Estado Bucket S3")

DynamoDB:

![3-dynamodb-empty.png](docs/images/3-dynamodb-empty.png?raw=true "Estado DynamoDB")

Tela de cadastro:

![4-cadastro.png](docs/images/4-cadastro.png?raw=true "Tela de cadastro")

Tela inicial com listagem dos cadastros:

![5-home-employees.png](docs/images/5-home-employees.png?raw=true "Tela inicial com listagem dos cadastros")

Listagem Bucket S3:

![6-bucket-s3-pictures.png](docs/images/6-bucket-s3-pictures.png?raw=true "Listagem Bucket S3")

Listagem DynamoDB:

![7-dynamodb-employees.png](docs/images/7-dynamodb-employees.png?raw=true "Listagem DynamoDB")
