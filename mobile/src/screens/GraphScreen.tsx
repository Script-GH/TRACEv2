import React, { useEffect, useState, useCallback } from 'react';
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity,
  ActivityIndicator, SafeAreaView, Dimensions,
} from 'react-native';
import Svg, { Circle, Line, Text as SvgText, Marker, Defs, Path } from 'react-native-svg';
import { API } from '../api/client';

const { width: SCREEN_W } = Dimensions.get('window');
const SVG_SIZE = SCREEN_W - 32;

type GraphNode = { id: string; type: string; label: string; x: number; y: number };
type GraphEdge = {
  id: string; from_node: string; to_node: string;
  relation: string; polarity: string; confidence: number;
};

interface GraphData {
  user_id: string;
  version: number;
  nodes: GraphNode[];
  edges: GraphEdge[];
}

function edgeColor(polarity: string): string {
  return polarity === 'POSITIVE' ? '#00D4AA' : '#FF4D6D';
}

function nodeColor(type: string): string {
  switch (type) {
    case 'EXERCISE': return '#3B82F6';
    case 'OUTCOME': return '#10B981';
    case 'STATE': return '#F59E0B';
    case 'NUTRITION': return '#8B5CF6';
    default: return '#6B7280';
  }
}

function scaleCoords(nodes: GraphNode[]): GraphNode[] {
  if (nodes.length === 0) return nodes;
  const pad = 60;
  const minX = Math.min(...nodes.map(n => n.x));
  const maxX = Math.max(...nodes.map(n => n.x));
  const minY = Math.min(...nodes.map(n => n.y));
  const maxY = Math.max(...nodes.map(n => n.y));
  const rangeX = maxX - minX || 1;
  const rangeY = maxY - minY || 1;
  return nodes.map(n => ({
    ...n,
    x: pad + ((n.x - minX) / rangeX) * (SVG_SIZE - pad * 2),
    y: pad + ((n.y - minY) / rangeY) * (SVG_SIZE - pad * 2),
  }));
}

function EdgeLine({ edge, nodes }: { edge: GraphEdge; nodes: GraphNode[] }) {
  const from = nodes.find(n => n.id === edge.from_node);
  const to = nodes.find(n => n.id === edge.to_node);
  if (!from || !to) return null;
  const color = edgeColor(edge.polarity);
  const opacity = 0.4 + edge.confidence * 0.6;
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const len = Math.sqrt(dx * dx + dy * dy) || 1;
  const r = 24;
  const x1 = from.x + (dx / len) * r;
  const y1 = from.y + (dy / len) * r;
  const x2 = to.x - (dx / len) * r;
  const y2 = to.y - (dy / len) * r;

  return <Line x1={x1} y1={y1} x2={x2} y2={y2} stroke={color} strokeWidth={1.5 + edge.confidence} strokeOpacity={opacity} />;
}

function NodeCircle({ node, onPress }: { node: GraphNode; onPress: (n: GraphNode) => void }) {
  const color = nodeColor(node.type);
  const label = node.label.length > 18 ? node.label.slice(0, 16) + '…' : node.label;
  return (
    <React.Fragment>
      <Circle
        cx={node.x} cy={node.y} r={24}
        fill={color} fillOpacity={0.15}
        stroke={color} strokeWidth={1.5}
        onPress={() => onPress(node)}
      />
      <SvgText
        x={node.x} y={node.y + 4}
        textAnchor="middle"
        fill={color} fontSize={9} fontWeight="600"
        onPress={() => onPress(node)}
      >
        {label}
      </SvgText>
    </React.Fragment>
  );
}

export default function GraphScreen() {
  const [graph, setGraph] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
  const [selectedEdges, setSelectedEdges] = useState<GraphEdge[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await API.getGraph();
      setGraph(data);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  function handleNodePress(node: GraphNode) {
    setSelectedNode(node);
    if (graph) {
      setSelectedEdges(graph.edges.filter(e => e.from_node === node.id || e.to_node === node.id));
    }
  }

  const scaledNodes = graph ? scaleCoords(graph.nodes) : [];

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.header}>
        <View>
          <Text style={styles.title}>Causal Graph</Text>
          {graph && <Text style={styles.subtitle}>v{graph.version} · {graph.nodes.length} nodes · {graph.edges.length} edges</Text>}
        </View>
        <TouchableOpacity onPress={load} style={styles.refreshBtn}>
          <Text style={styles.refreshText}>↺</Text>
        </TouchableOpacity>
      </View>

      {loading && <ActivityIndicator color="#00D4AA" style={styles.loader} />}

      {error && (
        <View style={styles.error}>
          <Text style={styles.errorText}>{error}</Text>
          <Text style={styles.errorHint}>Complete the interview first to build your graph.</Text>
        </View>
      )}

      {!loading && !error && graph && (
        <ScrollView contentContainerStyle={styles.scroll}>
          {/* SVG Graph */}
          <View style={styles.svgContainer}>
            <Svg width={SVG_SIZE} height={SVG_SIZE} style={styles.svg}>
              {/* Draw edges first (under nodes) */}
              {graph.edges.map(edge => (
                <EdgeLine key={edge.id} edge={edge} nodes={scaledNodes} />
              ))}
              {/* Draw nodes on top */}
              {scaledNodes.map(node => (
                <NodeCircle key={node.id} node={node} onPress={handleNodePress} />
              ))}
            </Svg>
          </View>

          {/* Legend */}
          <View style={styles.legend}>
            {[
              { color: '#3B82F6', label: 'Exercise' },
              { color: '#10B981', label: 'Outcome' },
              { color: '#F59E0B', label: 'State/Injury' },
              { color: '#6B7280', label: 'Pattern' },
            ].map(({ color, label }) => (
              <View key={label} style={styles.legendItem}>
                <View style={[styles.legendDot, { backgroundColor: color }]} />
                <Text style={styles.legendText}>{label}</Text>
              </View>
            ))}
          </View>
          <View style={styles.legend}>
            <View style={styles.legendItem}>
              <View style={[styles.legendLine, { backgroundColor: '#00D4AA' }]} />
              <Text style={styles.legendText}>Positive causal link</Text>
            </View>
            <View style={styles.legendItem}>
              <View style={[styles.legendLine, { backgroundColor: '#FF4D6D' }]} />
              <Text style={styles.legendText}>Negative causal link</Text>
            </View>
          </View>

          {/* Selected node detail */}
          {selectedNode && (
            <View style={styles.detail}>
              <Text style={styles.detailTitle}>{selectedNode.label}</Text>
              <Text style={styles.detailType}>{selectedNode.type}</Text>
              {selectedEdges.length > 0 && (
                <>
                  <Text style={styles.detailSectionLabel}>Causal links:</Text>
                  {selectedEdges.map(edge => {
                    const isFrom = edge.from_node === selectedNode.id;
                    const otherNode = graph.nodes.find(n => n.id === (isFrom ? edge.to_node : edge.from_node));
                    const arrow = isFrom ? '→' : '←';
                    const color = edgeColor(edge.polarity);
                    return (
                      <View key={edge.id} style={styles.edgeRow}>
                        <View style={[styles.edgeDot, { backgroundColor: color }]} />
                        <Text style={styles.edgeText}>
                          {arrow} {otherNode?.label ?? '?'}{' '}
                          <Text style={styles.edgeConf}>({edge.relation}, conf {edge.confidence.toFixed(2)})</Text>
                        </Text>
                      </View>
                    );
                  })}
                </>
              )}
            </View>
          )}
        </ScrollView>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: '#0A0A0A' },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingHorizontal: 20,
    paddingTop: 16,
    paddingBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#1A1A1A',
  },
  title: { color: '#EFEFEF', fontSize: 20, fontWeight: '700' },
  subtitle: { color: '#888', fontSize: 12, marginTop: 2 },
  refreshBtn: { padding: 8 },
  refreshText: { color: '#00D4AA', fontSize: 20 },
  loader: { flex: 1, justifyContent: 'center' },
  error: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 32 },
  errorText: { color: '#FF4D6D', fontSize: 16, marginBottom: 8, textAlign: 'center' },
  errorHint: { color: '#888', fontSize: 14, textAlign: 'center' },
  scroll: { paddingBottom: 40 },
  svgContainer: {
    alignItems: 'center',
    margin: 16,
    backgroundColor: '#111',
    borderRadius: 16,
    overflow: 'hidden',
  },
  svg: { backgroundColor: '#111' },
  legend: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 20,
    gap: 12,
    marginBottom: 4,
  },
  legendItem: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  legendDot: { width: 10, height: 10, borderRadius: 5 },
  legendLine: { width: 20, height: 3, borderRadius: 1.5 },
  legendText: { color: '#888', fontSize: 12 },
  detail: {
    margin: 16,
    backgroundColor: '#111',
    borderRadius: 16,
    padding: 16,
  },
  detailTitle: { color: '#EFEFEF', fontSize: 16, fontWeight: '700' },
  detailType: { color: '#888', fontSize: 12, marginTop: 2, marginBottom: 12 },
  detailSectionLabel: { color: '#888', fontSize: 12, marginBottom: 8 },
  edgeRow: { flexDirection: 'row', alignItems: 'flex-start', gap: 8, marginBottom: 6 },
  edgeDot: { width: 8, height: 8, borderRadius: 4, marginTop: 5 },
  edgeText: { color: '#EFEFEF', fontSize: 14, flex: 1 },
  edgeConf: { color: '#888' },
});
