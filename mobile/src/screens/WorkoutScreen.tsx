import React, { useEffect, useState, useCallback } from 'react';
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity,
  ActivityIndicator, SafeAreaView,
} from 'react-native';
import { API, Storage } from '../api/client';

interface Props {
  onReset?: () => void;
}

export default function WorkoutScreen({ onReset }: Props) {
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState<'pending' | 'ready' | 'error'>('pending');
  const [workoutData, setWorkoutData] = useState<any>(null);
  const [userId, setUserId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const uid = await Storage.getUserId();
      setUserId(uid);
      const data = await API.getWorkoutToday();
      setWorkoutData(data);
      setStatus('ready');
    } catch {
      // 202 Accepted means pending, other errors mean real problem
      setStatus('pending');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.header}>
        <View>
          <Text style={styles.title}>Today's Workout</Text>
          <Text style={styles.date}>{new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}</Text>
        </View>
        <View style={styles.headerBtns}>
          <TouchableOpacity onPress={load} style={styles.refreshBtn}>
            <Text style={styles.refreshText}>↺</Text>
          </TouchableOpacity>
          {onReset && (
            <TouchableOpacity
              onPress={async () => { await Storage.clear(); onReset(); }}
              style={styles.resetBtn}
            >
              <Text style={styles.resetText}>Reset Demo</Text>
            </TouchableOpacity>
          )}
        </View>
      </View>

      {loading && <ActivityIndicator color="#00D4AA" style={styles.loader} />}

      {!loading && (
        <ScrollView contentContainerStyle={styles.scroll}>
          {/* Status card */}
          <View style={[styles.card, styles.statusCard]}>
            <Text style={styles.cardIcon}>🧠</Text>
            <View style={styles.cardBody}>
              <Text style={styles.cardTitle}>Causal Graph Active</Text>
              <Text style={styles.cardSub}>
                Your personal causal model is ready. Workout generation uses your graph to explain every exercise choice.
              </Text>
            </View>
          </View>

          {workoutData && (
            <View style={styles.card}>
              <Text style={styles.cardTitle}>{workoutData.goal ?? 'Training Session'}</Text>
              <Text style={styles.cardSub}>{workoutData.message ?? 'Your workout is being prepared.'}</Text>
            </View>
          )}

          {/* Demo pipeline info */}
          <Text style={styles.sectionLabel}>Demo Status</Text>

          <PipelineStep
            step="1"
            label="Interview"
            description="Personal history collected"
            done={true}
          />
          <PipelineStep
            step="2"
            label="Causal Graph"
            description="Your body's causal model built from interview"
            done={true}
          />
          <PipelineStep
            step="3"
            label="Workout Generation"
            description="Daily plan derived from graph + MoE agents"
            done={false}
            next={true}
          />
          <PipelineStep
            step="4"
            label="Workout Logging"
            description="Log sessions to update your causal graph"
            done={false}
          />

          <View style={[styles.card, styles.graphCard]}>
            <Text style={styles.graphCardTitle}>How the graph drives your workout</Text>
            <Text style={styles.graphCardBody}>
              Each exercise in your plan will cite the causal edge that justified it.{'\n\n'}
              Example: "Barbell Row — because your graph shows Pull training CAUSED +15kg row strength (confidence 0.88)"
            </Text>
          </View>
        </ScrollView>
      )}
    </SafeAreaView>
  );
}

function PipelineStep({
  step, label, description, done, next,
}: { step: string; label: string; description: string; done: boolean; next?: boolean }) {
  return (
    <View style={[styles.pipelineStep, next && styles.pipelineStepNext]}>
      <View style={[styles.stepBadge, done ? styles.stepDone : next ? styles.stepNext : styles.stepPending]}>
        <Text style={styles.stepBadgeText}>{done ? '✓' : step}</Text>
      </View>
      <View style={styles.stepBody}>
        <Text style={[styles.stepLabel, done && styles.stepLabelDone]}>{label}</Text>
        <Text style={styles.stepDesc}>{description}</Text>
      </View>
    </View>
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
  date: { color: '#888', fontSize: 12, marginTop: 2 },
  headerBtns: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  refreshBtn: { padding: 8 },
  refreshText: { color: '#00D4AA', fontSize: 20 },
  resetBtn: { paddingHorizontal: 10, paddingVertical: 6, borderRadius: 8, borderWidth: 1, borderColor: '#333' },
  resetText: { color: '#666', fontSize: 12 },
  loader: { flex: 1 },
  scroll: { padding: 16, paddingBottom: 40 },
  card: {
    backgroundColor: '#111',
    borderRadius: 16,
    padding: 16,
    marginBottom: 12,
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: 12,
  },
  statusCard: { borderLeftWidth: 3, borderLeftColor: '#00D4AA' },
  cardIcon: { fontSize: 24 },
  cardBody: { flex: 1 },
  cardTitle: { color: '#EFEFEF', fontSize: 15, fontWeight: '600', marginBottom: 4 },
  cardSub: { color: '#888', fontSize: 13, lineHeight: 20 },
  sectionLabel: { color: '#888', fontSize: 12, fontWeight: '600', marginBottom: 8, marginTop: 8, textTransform: 'uppercase', letterSpacing: 1 },
  pipelineStep: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    backgroundColor: '#111',
    borderRadius: 12,
    padding: 14,
    marginBottom: 8,
    gap: 12,
  },
  pipelineStepNext: { borderWidth: 1, borderColor: '#00D4AA33' },
  stepBadge: {
    width: 32,
    height: 32,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stepDone: { backgroundColor: '#00D4AA' },
  stepNext: { backgroundColor: '#00D4AA22', borderWidth: 1, borderColor: '#00D4AA' },
  stepPending: { backgroundColor: '#1A1A1A' },
  stepBadgeText: { color: '#fff', fontSize: 13, fontWeight: '700' },
  stepBody: { flex: 1 },
  stepLabel: { color: '#888', fontSize: 14, fontWeight: '600', marginBottom: 2 },
  stepLabelDone: { color: '#EFEFEF' },
  stepDesc: { color: '#666', fontSize: 12 },
  graphCard: {
    marginTop: 8,
    flexDirection: 'column',
    borderLeftWidth: 3,
    borderLeftColor: '#3B82F6',
  },
  graphCardTitle: { color: '#3B82F6', fontSize: 14, fontWeight: '700', marginBottom: 8 },
  graphCardBody: { color: '#888', fontSize: 13, lineHeight: 20 },
});
