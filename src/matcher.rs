use std::vec::Vec;

#[derive(Clone)]
#[derive(Debug)]
struct Mid {
    chars: Vec<(u8, Node)>,
    number: Option<Box<Node>>,
}

type Action = u8;

#[derive(Debug)]
#[derive(Clone)]
enum Node {
    Accept(Action),
    Mid(Mid),
}

fn new(pat: &mut Iterator<Item = u8>, action: Action) -> Node {
    match pat.next() {
        None => Node::Accept(action),
        Some(c) => {
            Node::Mid(Mid {
                chars: vec![(c, new(pat, action))],
                number: None,
            })
        }
    }
}

impl Node {
    fn add(&mut self, pat: &mut Iterator<Item = u8>, action: Action) {
        match pat.next() {
            None => {
                match *self {
                    Node::Accept(_) => panic!("already matched"),
                    Node::Mid(ref mid) if mid.chars.len() > 0 => panic!("longer match exists"),
                    _ => {}
                }
                *self = Node::Accept(action);
            }
            Some(c) => {
                match *self {
                    Node::Accept(_) => panic!("shorter match exists"),
                    Node::Mid(ref mut mid) => {
                        for &mut (ref m, ref mut n) in &mut mid.chars {
                            if *m == c {
                                n.add(pat, action);
                                return;
                            }
                        }
                        mid.chars.push((c, new(pat, action)));
                    }
                }
            }
        }
    }
}


fn test(start: &Node, input: &mut Iterator<Item = u8>) -> Action {
    let mut node = start;
    loop {
        let mid = match *node {
            Node::Accept(a) => return a,
            Node::Mid(ref mid) => mid,
        };
        let b = input.next().unwrap();
        match mid.chars.iter().find(|&&(m, _)| m == b) {
            None => panic!("no match"),
            Some(&(_, ref n)) => node = n,
        }
    }
}

fn print_indent(indent: usize) {
    for _ in 0..indent {
        print!("  ");
    }
}

fn dump(node: &Node, indent: usize) {
    match *node {
        Node::Accept(a) => {
            print_indent(indent);
            println!("accept {}", a)
        }
        Node::Mid(ref mid) => {
            for &(ref m, ref n) in mid.chars.iter() {
                print_indent(indent);
                println!("{} =>", *m as char);
                dump(n, indent + 1);
            }
        }
    }
}

pub fn main() {
    let mut n = new(&mut "\x1b[?%h".bytes(), 3);
    n.add(&mut "bc".bytes(), 4);
    n.add(&mut "bd".bytes(), 7);

    dump(&n, 0);
    println!("match {}", test(&n, &mut "bcd".bytes()));
}
