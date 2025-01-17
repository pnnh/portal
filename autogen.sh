oapi-codegen -package gen -generate types quark/domain/taylorpetstore.yaml > gen/taylorpetstore.types.gen.go
oapi-codegen -package gen -generate spec quark/domain/taylorpetstore.yaml > gen/taylorpetstore.spec.gen.go
oapi-codegen -package gen -generate gin quark/domain/taylorpetstore.yaml > gen/taylorpetstore.ginserver.gen.go
oapi-codegen -package gen -generate client quark/domain/taylorpetstore.yaml > gen/taylorpetstore.client.gen.go