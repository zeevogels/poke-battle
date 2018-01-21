package calculator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
)

// Calculator knows pokemon
type Calculator struct {
	pokemon     map[int]*PokeBase
	moves       map[int]*Move
	typeEffects map[string]float64
}

// NewCalculator returns a new loaded calculator
func NewCalculator() *Calculator {
	calc := &Calculator{}
	calc.readPokemon()
	calc.readMoves()
	calc.readEffects()
	return calc
}
func (c *Calculator) readPokemon() {

	var pokes Pokebases
	err := readJSON("calculator/pokemon.json", &pokes)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("read pokemon: ", len(pokes))

	pokesList := []string{}
	for i := 0; i < len(pokes); i++ {
		pokesList = append(pokesList, pokes[i].Name)
	}
	hotEncodedList := c.hotEncodeList(pokesList)

	pokemap := make(map[int]*PokeBase)
	for i := 0; i < len(pokes); i++ {
		pokes[i].HotEncoding = hotEncodedList[pokes[i].Name]
		poke := &pokes[i]
		//log.Println(poke.Number)
		pokemap[poke.Number] = poke
	}
	c.pokemon = pokemap
}
func (c *Calculator) readMoves() {
	var moves AllMoves
	err := readJSON("calculator/moves.json", &moves)
	if err != nil {
		log.Fatalln(err)
	}

	movesList := []string{}
	for i := 0; i < len(moves); i++ {
		movesList = append(movesList, moves[i].Name)
	}
	hotEncodedList := c.hotEncodeList(movesList)

	movemap := make(map[int]*Move)
	for i := 0; i < len(moves); i++ {
		//log.Println(&moves[i])
		moves[i].HotEncoding = hotEncodedList[moves[i].Name]
		movemap[i+1] = &moves[i]
		movemap[i+1].MoveID = i + 1
	}
	c.moves = movemap

}
func (c *Calculator) readEffects() {
	var typeEffects AllEffectivenesses
	err := readJSON("calculator/type-effectiveness.json", &typeEffects)
	if err != nil {
		log.Fatalln(err)
	}
	typemap := make(map[string]float64)
	for i := 0; i < len(typeEffects); i++ {
		//log.Println(fmt.Sprintf("%s-%s", typeEffects[i].Attack, typeEffects[i].Defend))
		typeformat := fmt.Sprintf("%s-%s", typeEffects[i].Attack, typeEffects[i].Defend)
		typemap[typeformat] = typeEffects[i].Effectiveness
	}
	c.typeEffects = typemap
}

// GetRandomSpecificPokemon generates a given pokemon with given level
func (c *Calculator) GetRandomSpecificPokemon(pokenumber, level int) *Pokemon {
	log.Println("generating random pokemon")

	//get pokemon base
	base := c.pokemon[pokenumber]
	log.Println("its a ", base.Name)

	//get random level between 1 and 99
	log.Println("with level: ", level)

	//generate a random ABLE moveset for the give pokemon
	moveset := c.generateMoveset(base, level)
	for i := 0; i < len(moveset); i++ {
		log.Println("has move: ", moveset[i].Name)
	}

	//calculate stats for base, level and generated IV's and EV's
	//calculate random EV's
	var attackEV, defenseEV, speedEV, specialEV, HPEV, max int
	if level < 5 {
		max = level
	} else {
		max = ((level - 3) * (level - 3)) + 1
	}
	defeated := random(0, max)
	log.Printf("This pokemon has defeated %d pokemons already", defeated)
	for index := 0; index < defeated; index++ {
		randdefeated := c.pokemon[random(1, 151)]
		attackEV += randdefeated.BaseStats.Attack
		defenseEV += randdefeated.BaseStats.Defense
		speedEV += randdefeated.BaseStats.Speed
		specialEV += randdefeated.BaseStats.Special
		HPEV += randdefeated.BaseStats.Hp
	}

	log.Printf("EV attack: %d, defense: %d, hp: %d, speed: %d, special: %d", attackEV, defenseEV, HPEV, speedEV, specialEV)
	if HPEV > 65535 {
		HPEV = 65535
	}

	if defenseEV > 65535 {
		defenseEV = 65535
	}

	if attackEV > 65535 {
		attackEV = 65535
	}
	if specialEV > 65535 {
		specialEV = 65535
	}

	if speedEV > 65535 {
		speedEV = 65535
	}
	//calculate IV's
	attackIV := random(0, 15)
	defenseIV := random(0, 15)
	speedIV := random(0, 15)
	specialIV := random(0, 15)
	//hpIV is calculated
	var hpIV int
	if !even(attackIV) {
		hpIV += 8
	} // 1111 = 8 + 4 + 2 + 1 = 15
	if !even(defenseIV) {
		hpIV += 4
	}
	if !even(speedIV) {
		hpIV += 2
	}
	if !even(specialIV) {
		hpIV++
	}
	log.Printf("IV attack: %d, defense: %d, hp: %d, speed: %d, special: %d", attackIV, defenseIV, hpIV, speedIV, specialIV)
	stats := &BaseStats{
		Hp:      c.calculateHP(base.BaseStats.Hp, hpIV, HPEV, level),
		Attack:  c.calculateOtherStat(base.BaseStats.Attack, attackIV, attackEV, level),
		Defense: c.calculateOtherStat(base.BaseStats.Defense, defenseIV, defenseEV, level),
		Speed:   c.calculateOtherStat(base.BaseStats.Speed, speedIV, speedEV, level),
		Special: c.calculateOtherStat(base.BaseStats.Special, specialIV, specialEV, level),
	}

	log.Printf("attack: %d, defense: %d, hp: %d, speed: %d, special: %d", stats.Attack, stats.Defense, stats.Hp, stats.Speed, stats.Special)

	pokemon := &Pokemon{
		base:    base,
		level:   level,
		moveset: moveset,
		stats:   *stats,
		status:  Fit,
	}

	return pokemon
}

// GetRandomPokemon generates a random pokemon with a random level and a random possible moveset
func (c *Calculator) GetRandomPokemon() *Pokemon {
	log.Println("generating random pokemon")

	//get pokemon base
	base := c.pokemon[random(1, 151)]
	log.Println("its a ", base.Name)

	//get random level between 1 and 99
	level := random(MINLEVEL, MAXLEVEL)
	log.Println("with level: ", level)

	//generate a random ABLE moveset for the give pokemon
	moveset := c.generateMoveset(base, level)
	for i := 0; i < len(moveset); i++ {
		log.Println("has move: ", moveset[i].Name)
	}

	//calculate stats for base, level and generated IV's and EV's
	//calculate random EV's
	var attackEV, defenseEV, speedEV, specialEV, HPEV, max int
	if level < 5 {
		max = level
	} else {
		max = ((level - 3) * (level - 3)) + 1
	}
	max /= 3
	defeated := random(0, max)
	log.Printf("This pokemon has defeated %d pokemons already", defeated)
	for index := 0; index < defeated; index++ {
		randdefeated := c.pokemon[random(1, 151)]
		attackEV += randdefeated.BaseStats.Attack
		defenseEV += randdefeated.BaseStats.Defense
		speedEV += randdefeated.BaseStats.Speed
		specialEV += randdefeated.BaseStats.Special
		HPEV += randdefeated.BaseStats.Hp
	}
	if HPEV > 65535 {
		HPEV = 65535
	}

	if defenseEV > 65535 {
		defenseEV = 65535
	}

	if attackEV > 65535 {
		attackEV = 65535
	}
	if specialEV > 65535 {
		specialEV = 65535
	}

	if speedEV > 65535 {
		speedEV = 65535
	}
	//calculate IV's
	attackIV := random(0, 15)
	defenseIV := random(0, 15)
	speedIV := random(0, 15)
	specialIV := random(0, 15)
	//hpIV is calculated
	var hpIV int
	if !even(attackIV) {
		hpIV += 8
	} // 1111 = 8 + 4 + 2 + 1 = 15
	if !even(defenseIV) {
		hpIV += 4
	}
	if !even(speedIV) {
		hpIV += 2
	}
	if !even(specialIV) {
		hpIV++
	}

	stats := &BaseStats{
		Hp:      c.calculateHP(base.BaseStats.Hp, hpIV, HPEV, level),
		Attack:  c.calculateOtherStat(base.BaseStats.Attack, attackIV, attackEV, level),
		Defense: c.calculateOtherStat(base.BaseStats.Defense, defenseIV, defenseEV, level),
		Speed:   c.calculateOtherStat(base.BaseStats.Speed, speedIV, speedEV, level),
		Special: c.calculateOtherStat(base.BaseStats.Special, specialIV, specialEV, level),
	}

	log.Printf("attack: %d, defense: %d, hp: %d, speed: %d, special: %d", stats.Attack, stats.Defense, stats.Hp, stats.Speed, stats.Special)

	pokemon := &Pokemon{
		base:    base,
		level:   level,
		moveset: moveset,
		stats:   *stats,
		status:  Fit,
	}

	return pokemon
}
func (c *Calculator) generateMoveset(poke *PokeBase, level int) [4]*Move {
	var moves [4]*Move

	for i := 0; i < 4; i++ {
		random := random(1, len(c.moves))
		moves[i] = c.moves[random]
	}
	return moves
}

// Returns whether poke1 attacks first or not
func (c *Calculator) poke1First(poke1Speed, poke2Speed, poke1Priority, poke2Priority int) bool{

		// The pokemon that selected the move with the highest priority will attack first
		if poke1Priority > poke2Priority {
			return true
		}
		if poke2Priority > poke1Priority{
			return false
		}
		//  if both moves have the same priority, the pokemon with the higher speed will attack first
		if poke1Speed > poke2Speed {
			return true
		}
		if poke2Speed > poke1Speed {
			return false
		}
		// if both moves have the same priority and both pokemon have the same speed
		// it is determined randomly who will attack first
		if random(0, 1) == 0 {
			return true
		}

		return false
}
// Fight simulates a fight between poke1 and poke2
func (c *Calculator) Fight(poke1 *Pokemon, poke2 *Pokemon) *Pokemon {

	// As long as one of the pokemon hasnt reached 0 HP the fight isnt over yet
	for {
		//let poke1 and poke2 choose a random move before fighting
		poke1move := poke1.moveset[random(0, 3)]
		poke2move := poke2.moveset[random(0, 3)]

		//it is then decided in this method which pokemon will attack first
		poke1first := c.poke1First(poke1.Speed(), poke2.Speed(),poke1move.Priority, poke2move.Priority)

		//the effectiveness of each move is decided from the type of the move and the type of the defending pokemon
		effectiveness1 := c.getTypeEffectiveness(poke1, poke2move)
		effectiveness2 := c.getTypeEffectiveness(poke2, poke1move)
		//log.Println(effectiveness2, " / ", poke2.base.Types, poke1move.MoveType)
		//log.Println(effectiveness1, " / ", poke1.base.Types, poke2move.MoveType)
		
		// attacking happens here
		// if an attack returns true it means the defending pokemon died and the attacking pokemon wil be returned as winner
		// if a pokemon misses its attack, nothing happens
		if poke1first {
			// Pokemon 1 attacks first
			if c.isPokemonAbleToAttack(poke1){
				if c.doesMoveHit(poke1move,poke2.accuracyBoost,poke1.evasivenessBoost){
					// Pokemon 1's attack strikes
					if c.Attack(poke1move, poke2, poke1, effectiveness2){
						return poke1
					}
				}else{
					// Pokemon 1's attack misses
					log.Println("Attack Missed")
				}
			}
			// Pokemon 2 attacks second
			if c.isPokemonAbleToAttack(poke2){
				if c.doesMoveHit(poke2move, poke1.accuracyBoost, poke2.evasivenessBoost){
					if c.Attack(poke2move, poke1, poke2, effectiveness1){
						return poke2
					}
				}else{
					// Pokemon 2's attack misses
					log.Println("Attack Missed")
				}
			}
		} else {
			// Pokemon 2 attacks first
			if c.isPokemonAbleToAttack(poke2){
				if c.doesMoveHit(poke2move, poke1.accuracyBoost, poke2.evasivenessBoost){
					if c.Attack(poke2move, poke1, poke2, effectiveness1){
						return poke2
					}
				}else{
					// Pokemon 2's attack misses
					log.Println("Attack Missed")
				}
			}
			// Pokemon 1 attacks second
			if c.isPokemonAbleToAttack(poke1){
				if c.doesMoveHit(poke1move,poke2.accuracyBoost,poke1.evasivenessBoost){
					// Pokemon 1's attack strikes
					if c.Attack(poke1move, poke2, poke1, effectiveness2){
						return poke1
					}
				}else{
					// Pokemon 1's attack misses
					log.Println("Attack Missed")
				}
			}
		}
	}
}
// Effect handles the effect that comes with an attack causing one of the following effects:
// Effects handled: 
func (c *Calculator) Effect(movename string, poke, enemy *Pokemon){

}
// Attack attacks and returns true if the pokemon dies
func (c *Calculator) Attack(enemyMove *Move, poke, enemy *Pokemon, effectiveness float64) bool {
	log.Printf("%s uses %s", enemy.base.Name, enemyMove.Name)
	var attack, defense int
	damage := (2 * enemy.level) / 5
	//log.Println("step 1: ", damage)
	damage += 2
	//log.Println("step 2: ", damage)
	switch enemyMove.Category {
	case "physical":
		attack = enemy.stats.Attack
		defense = poke.stats.Defense
	case "special":
		attack = enemy.stats.Special
		defense = poke.stats.Special
	case "status":
		attack = 4
		defense = 2
	}
	//log.Println("base power:", enemyMove.Power)
	//log.Println("attacck: ", attack)
	//log.Println("defe: ", defense)
	ada := float64(attack) / float64(defense)
	log.Println("Power: ", enemyMove.Power)
	damage *= int(float64(enemyMove.Power) * ada)
	//log.Println("step 3: ", damage)
	damage /= 50
	//log.Println("step 4: ", damage)
	damage += 2
	//log.Println("step 5: ", damage)
	//modifier = random * stab * type effect

	//random is between 0.85 and 1
	Mrandom := float64(random(218, 255))
	Mrandom /= 255

	//stab
	Mstab := 1.0
	for i := 0; i < len(enemy.base.Types); i++ {
		if enemy.base.Types[i] == string(enemyMove.MoveType) {
			Mstab = 1.5
		}
	}
	log.Println("Stab: ", Mstab)
	// type effect was given in param
	log.Println("effectivness: ", effectiveness)
	// So modifier is calculated
	modifier := Mrandom * Mstab *effectiveness
	log.Println("modifier: ", modifier)

	//modified damage calculation
	damage = int(modifier * float64(damage))
	//log.Println("step 6: ", damage)
	log.Println("actual damage: ", damage)
	poke.stats.Hp -= damage
	if poke.stats.Hp <= 0 {
		log.Printf("%s dies", poke.base.Name)
		return true
	}
	log.Printf("%s has %d hp left", poke.base.Name, poke.stats.Hp)
	return false
}
func (c *Calculator) getTypeEffectiveness(poke *Pokemon, move *Move) float64 {
	effectiveness := c.typeEffects[string(move.MoveType)+"-"+poke.base.Types[0]]

	if len(poke.base.Types) > 1 {
		effectiveness *= c.typeEffects[string(move.MoveType)+"-"+poke.base.Types[1]]
	}

	return effectiveness
}
func (c *Calculator) isPokemonAbleToAttack(poke *Pokemon) bool {
	if poke.status == Asleep || poke.status == FrozenSolid || poke.status == Flinched{
		return false
	}
	if poke.status == Paralyzed{
		if random(1,100) > 75 {
			return false
		}
	}
	return true
}
func (c *Calculator) doesMoveHit(move *Move, selfAccuracyStage, enemyEvasionStage int) bool {
	//always hits
	if move.Name == "swift" {
		return true
	}
	// 4, 1
	selfMultiplier := c.getStageMultiplier(true, selfAccuracyStage)
	enemyMultiplier := c.getStageMultiplier(false, enemyEvasionStage)
	
	//1/256 glitch if multipliers are unchanged 
	if selfMultiplier == 1 && enemyMultiplier == 1 {
		if random(1,256) == 256{
			log.Println("glitch")
			return false
		}
	}

	hitratio := (selfMultiplier * enemyMultiplier) * move.Accuracy
	if hitratio >= 1 {
		return true
	}

	chances := random(1, 100)
	if (hitratio*100)-float64(chances) > 0 {
		return true
	}
	return false
}
func (c *Calculator) getStageMultiplier(accuracy bool, stage int) float64 {
	stageMultipliers := []float64{0.25, 0.28, 0.33, 0.4, 0.5, 0.66, 1, 1.5, 2, 2.5, 3, 3.5, 4}
	// if its not accuracy, its evasion
	if !accuracy {
		stage -= 6
		// stage should now be between -12 and 0
		newStage := math.Abs(float64(stage))
		return stageMultipliers[int(newStage)]
	}
	// stage should be between 0 and 12
	return stageMultipliers[stage+6]
}
func (c *Calculator) calculateHP(baseHP, hpIV, hpEV, level int) int {
	//stats are rounded down if decimal
	sqrtEV := math.Sqrt(float64(hpEV)) / 4
	baseIV := float64(baseHP + hpIV)
	hp := (((baseIV*2 + sqrtEV) * float64(level)) / 100) + float64(level) + 10
	return int(hp)
}
func (c *Calculator) calculateOtherStat(base, iv, ev, level int) int {
	//stats are rounded down if decimal
	baseIV := float64(base + iv)
	sqrtEV := math.Sqrt(float64(ev)) / 4
	//this calculation is rounded down for any stat except HP
	stat := int(baseIV*2 + sqrtEV)
	stat = (stat * level) / 100
	return stat + 5
}

func (c *Calculator) hotEncodeList(list []string) map[string]string {
	encodedList := make(map[string]string, len(list))

	for j := 0; j < len(list); j++ {
		element := "("
		for i := 0; i < len(list); i++ {
			if i != j {
				element += "0"
			} else {
				element += "1"
			}

			if i != (len(list) - 1) {
				element += ","
			}
		}
		element += ")"
		encodedList[list[j]] = element
	}
	return encodedList
}
func readJSON(uri string, sort interface{}) error {
	file, err := ioutil.ReadFile(uri)
	if err != nil {
		return err
	}
	//m := new(Dispatch)
	//var m interface{}
	err = json.Unmarshal(file, sort)
	if err != nil {
		return err
	}
	return nil
}
func random(min, max int) int {
	if (max - min) <= 0 {
		return 1
	}
	return rand.Intn(max-min) + min
}
func even(number int) bool {
	return number%2 == 0
}
