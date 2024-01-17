package constants

const (
	DefaultWithFlushNum = 5000
)

const (
	// KV Bucket
	BucketHeightToBlockHeader                           = "HEIGHT_TO_BLOCK_HEADER"
	BucketHeightToLastSequenceNumber                    = "HEIGHT_TO_LAST_SEQUENCE_NUMBER"
	BucketInscriptionIdToSequenceNumber                 = "INSCRIPTION_ID_TO_SEQUENCE_NUMBER"
	BucketInscriptionNumberToSequenceNumber             = "INSCRIPTION_NUMBER_TO_SEQUENCE_NUMBER"
	BucketOutpointToSatRanges                           = "OUTPOINT_TO_SAT_RANGES"
	BucketOutpointToValue                               = "OUTPOINT_TO_VALUE"
	BucketOutpointToInscriptions                        = "OUTPOINT_TO_INSCRIPTIONS"
	BucketSatToSatpoint                                 = "SAT_TO_SATPOINT"
	BucketSequenceNumberToInscriptionEntry              = "SEQUENCE_NUMBER_TO_INSCRIPTION_ENTRY"
	BucketSequenceNumberToSatpoint                      = "SEQUENCE_NUMBER_TO_SATPOINT"
	BucketStatisticToCount                              = "STATISTIC_TO_COUNT"
	BucketTransactionIdToTransaction                    = "TRANSACTION_ID_TO_TRANSACTION"
	BucketWriteTransactionStartingBlockCountToTimestamp = "WRITE_TRANSACTION_STARTING_BLOCK_COUNT_TO_TIMESTAMP"

	// Set Bucket
	BucketSatpointToSequenceNumber = "SATPOINT_TO_SEQUENCE_NUMBER"
	BucketSatToSequenceNumber      = "SAT_TO_SEQUENCE_NUMBER"

	BucketKeyList = "KEY_LIST"
)

var KVBuckets = []string{
	BucketHeightToBlockHeader,
	BucketHeightToLastSequenceNumber,
	BucketInscriptionIdToSequenceNumber,
	BucketInscriptionNumberToSequenceNumber,
	BucketOutpointToSatRanges,
	BucketOutpointToValue,
	BucketOutpointToInscriptions,
	BucketSatToSatpoint,
	BucketSequenceNumberToInscriptionEntry,
	BucketSequenceNumberToSatpoint,
	BucketStatisticToCount,
	BucketTransactionIdToTransaction,
	BucketWriteTransactionStartingBlockCountToTimestamp,
}

var SetBuckets = []string{
	BucketSatpointToSequenceNumber,
	BucketSatToSequenceNumber,
}

type Statistic string

var Statistics = []Statistic{
	StatisticSchema,
	StatisticBlessedInscriptions,
	StatisticCommits,
	StatisticCursedInscriptions,
	StatisticIndexRunes,
	StatisticIndexSats,
	StatisticLostSats,
	StatisticOutputsTraversed,
	StatisticReservedRunes,
	StatisticRunes,
	StatisticSatRanges,
	StatisticUnboundInscriptions,
	StatisticIndexTransactions,
}

const (
	StatisticSchema              Statistic = "0"
	StatisticBlessedInscriptions Statistic = "1"
	StatisticCommits             Statistic = "2"
	StatisticCursedInscriptions  Statistic = "3"
	StatisticIndexRunes          Statistic = "4"
	StatisticIndexSats           Statistic = "5"
	StatisticLostSats            Statistic = "6"
	StatisticOutputsTraversed    Statistic = "7"
	StatisticReservedRunes       Statistic = "8"
	StatisticRunes               Statistic = "9"
	StatisticSatRanges           Statistic = "10"
	StatisticUnboundInscriptions Statistic = "11"
	StatisticIndexTransactions   Statistic = "12"
)
