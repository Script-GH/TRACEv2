import React, { useState, useRef, useEffect } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, FlatList,
  KeyboardAvoidingView, Platform, StyleSheet, ActivityIndicator,
  SafeAreaView, Animated,
} from 'react-native';
import { API, Storage } from '../api/client';

interface Message {
  id: string;
  role: 'user' | 'agent';
  content: string;
}

interface Props {
  onComplete: () => void;
}

export default function InterviewScreen({ onComplete }: Props) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [isComplete, setIsComplete] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const listRef = useRef<FlatList>(null);
  const progressAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    init();
  }, []);

  useEffect(() => {
    Animated.timing(progressAnim, {
      toValue: progress,
      duration: 600,
      useNativeDriver: false,
    }).start();
  }, [progress]);

  async function init() {
    setLoading(true);
    try {
      // Token is guaranteed set by App.tsx before InterviewScreen mounts.
      const { session_id, opening_message } = await API.startInterview();
      setSessionId(session_id);
      addMessage('agent', opening_message);
    } catch (e: any) {
      addMessage('agent', 'Failed to connect to TRACE. Check your connection and try again.');
    } finally {
      setLoading(false);
    }
  }

  function addMessage(role: 'user' | 'agent', content: string) {
    const msg: Message = { id: Date.now().toString() + Math.random(), role, content };
    setMessages(prev => [...prev, msg]);
    setTimeout(() => listRef.current?.scrollToEnd({ animated: true }), 100);
  }

  async function sendMessage() {
    if (!input.trim() || !sessionId || loading) return;
    const text = input.trim();
    setInput('');
    addMessage('user', text);
    setLoading(true);

    try {
      const result = await API.sendTurn(sessionId, text);
      addMessage('agent', result.agent_reply);
      setProgress(result.progress_pct);
      if (result.is_complete) {
        setIsComplete(true);
      }
    } catch (e) {
      addMessage('agent', 'Something went wrong. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  async function handleConfirm() {
    if (!sessionId || confirming) return;
    setConfirming(true);
    addMessage('agent', 'Saving your profile and building your causal graph...');
    try {
      await API.confirmInterview(sessionId);
      await Storage.setInterviewDone();
      addMessage('agent', 'Your personal causal graph is built! Head to the Graph tab to explore it.');
      setTimeout(onComplete, 1500);
    } catch (e: any) {
      addMessage('agent', 'Error: ' + e.message);
      setConfirming(false);
    }
  }

  const progressWidth = progressAnim.interpolate({
    inputRange: [0, 100],
    outputRange: ['0%', '100%'],
  });

  return (
    <SafeAreaView style={styles.safe}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>TRACE Onboarding</Text>
        <View style={styles.progressBar}>
          <Animated.View style={[styles.progressFill, { width: progressWidth }]} />
        </View>
        <Text style={styles.progressText}>{progress}% complete</Text>
      </View>

      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        keyboardVerticalOffset={0}
      >
        <FlatList
          ref={listRef}
          data={messages}
          keyExtractor={m => m.id}
          style={styles.messages}
          contentContainerStyle={styles.messagesContent}
          renderItem={({ item }) => (
            <View style={[styles.bubble, item.role === 'user' ? styles.userBubble : styles.agentBubble]}>
              <Text style={[styles.bubbleText, item.role === 'user' ? styles.userText : styles.agentText]}>
                {item.content}
              </Text>
            </View>
          )}
          ListFooterComponent={loading ? <ActivityIndicator style={styles.typing} color="#00D4AA" /> : null}
        />

        {isComplete && (
          <TouchableOpacity
            style={[styles.confirmBtn, confirming && styles.confirmBtnDisabled]}
            onPress={handleConfirm}
            disabled={confirming}
          >
            {confirming
              ? <ActivityIndicator color="#fff" />
              : <Text style={styles.confirmText}>Build My Causal Graph →</Text>
            }
          </TouchableOpacity>
        )}

        <View style={styles.inputRow}>
          <TextInput
            style={styles.input}
            value={input}
            onChangeText={setInput}
            placeholder="Type your answer..."
            placeholderTextColor="#888"
            multiline
            maxLength={500}
            editable={!isComplete && !loading}
            onSubmitEditing={sendMessage}
            returnKeyType="send"
          />
          <TouchableOpacity
            style={[styles.sendBtn, (!input.trim() || loading || isComplete) && styles.sendBtnDisabled]}
            onPress={sendMessage}
            disabled={!input.trim() || loading || isComplete}
          >
            <Text style={styles.sendText}>↑</Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: '#0A0A0A' },
  flex: { flex: 1 },
  header: {
    paddingHorizontal: 20,
    paddingTop: 12,
    paddingBottom: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#1A1A1A',
  },
  headerTitle: { color: '#00D4AA', fontSize: 16, fontWeight: '700', marginBottom: 8 },
  progressBar: {
    height: 4,
    backgroundColor: '#1A1A1A',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#00D4AA',
    borderRadius: 2,
  },
  progressText: { color: '#888', fontSize: 12, marginTop: 4 },
  messages: { flex: 1 },
  messagesContent: { paddingHorizontal: 16, paddingVertical: 12 },
  bubble: {
    maxWidth: '85%',
    padding: 12,
    borderRadius: 16,
    marginVertical: 4,
  },
  userBubble: {
    backgroundColor: '#00D4AA',
    alignSelf: 'flex-end',
    borderBottomRightRadius: 4,
  },
  agentBubble: {
    backgroundColor: '#1A1A1A',
    alignSelf: 'flex-start',
    borderBottomLeftRadius: 4,
  },
  bubbleText: { fontSize: 15, lineHeight: 22 },
  userText: { color: '#0A0A0A' },
  agentText: { color: '#EFEFEF' },
  typing: { marginVertical: 8, alignSelf: 'flex-start', marginLeft: 16 },
  confirmBtn: {
    margin: 16,
    backgroundColor: '#00D4AA',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
  },
  confirmBtnDisabled: { opacity: 0.6 },
  confirmText: { color: '#0A0A0A', fontSize: 16, fontWeight: '700' },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderTopWidth: 1,
    borderTopColor: '#1A1A1A',
    gap: 8,
  },
  input: {
    flex: 1,
    backgroundColor: '#1A1A1A',
    color: '#EFEFEF',
    borderRadius: 12,
    paddingHorizontal: 14,
    paddingVertical: 10,
    fontSize: 15,
    maxHeight: 120,
  },
  sendBtn: {
    width: 40,
    height: 40,
    backgroundColor: '#00D4AA',
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
  },
  sendBtnDisabled: { backgroundColor: '#1A1A1A' },
  sendText: { color: '#0A0A0A', fontSize: 20, fontWeight: '700' },
});
