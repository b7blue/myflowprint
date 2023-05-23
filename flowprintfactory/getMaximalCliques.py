import networkx as nx
import sys

edges = []
G = nx.Graph()
f = open('extension.txt','r')
raw = f.read()
rawedges = raw.split('.')
for rawedge in rawedges:
    vertex = rawedge.split(',')
    edges.append((int(vertex[0]),int(vertex[1])))
G.add_edges_from(edges)
res = nx.find_cliques(G)

for maximal_cliques in res:
    print(maximal_cliques)




