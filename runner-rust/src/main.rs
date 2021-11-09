use actix_web::{web, App, HttpServer, Responder};
use rhai::{Engine, Scope};
use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct Body {
    name: String,
}

struct Controller<'a> {
    engine: &'a Engine,
}

impl<'a> Controller<'a> {
    fn new(engine: &'a Engine) -> Self {
        Self { engine }
    }
}

async fn run_strategy(body: web::Json<Body>) -> impl Responder {
    let engine = Engine::new();
    let ast = engine.compile_file((&body.name).into()).unwrap();
    let mut scope = Scope::new();

    let res: i32 = engine.call_fn(&mut scope, &ast, "sum", (2, 3)).unwrap();

    return res.to_string();
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new().service(web::resource("/strategy/run").route(web::post().to(run_strategy)))
    })
    .bind("127.0.0.1:3000")?
    .run()
    .await
}
